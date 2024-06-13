    import React, { useState, useEffect } from 'react';
    import './summary.scss';
    import { FeedbackComponent } from '../feedback/feedback';
    import { ResourceCards } from '../resourcecard/resourcecard';
    import * as Const from '../../shared/constants';
    import Moment from 'react-moment';

    function formatTimeDifference (startTime: string, endTime: string)  {
        const diffInSeconds = Math.floor((new Date(endTime).getTime() - new Date(startTime).getTime()) / 1000);
        const seconds = diffInSeconds % 60;
        return `${seconds}s`;
    };

    export const Summary = ({ application, resource , }) => {
        const [obj, setObj] = useState(null);

        //const [disableButton, setDisableButton] = useState(true);
        const {metadata, spec, status} = application;
        const applicationName = metadata?.name || '';
        const applicationNamespace = metadata?.namespace || '';
        const destNamespace = spec?.destination?.namespace || '';
        const isHealthy = status?.health?.status !== "Healthy" || status?.conditions && status.conditions.length;



        useEffect(() => {
            fetchStatus("create-genai");
        }, [status?.health?.status !== "Healthy",status?.conditions && status.conditions]);

        const fetchStatus = async (requestType: string) => {
            try {
                let jsonData = null;
                do {
                    const statusResponse = await fetch(Const.APIs.fetchGenAIStatus(applicationName, applicationNamespace, destNamespace));
                    if (statusResponse.ok) {
                        const data = await statusResponse.json();
                        jsonData = typeof data.manifest === "string" ? JSON.parse(data.manifest) : data.manifest;
                    } else {
                        console.log('GenAI does not exist');
                        if(requestType != ""){
                            await GenAIResource(requestType);
                        }
                    }
                    setObj(jsonData);
                    await new Promise(resolve => setTimeout(resolve, 5000));
                } while (jsonData?.status?.phase !== "completed" && jsonData?.status?.phase !== "error");

            } catch (error) {
                console.error('Error fetching GenAI status:', error);
            }
        };

        const patchAnnotation = async () => {
            const url = Const.APIs.patchAnnotation(applicationName, applicationNamespace, destNamespace);
            const patch = JSON.stringify({
                "metadata": {
                    "annotations": {
                        "argosupport.argoproj.extensions.io/genai": `${JSON.stringify(application.status)}`
                    }
                }
            });
            try {
                const response = await fetch(url, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(patch)
                });

                if (!response.ok) throw new Error('error patching the genai resource');

                const patchedData = await response.json();
                await new Promise(resolve => setTimeout(resolve, 800));
                console.log('Patch request successful:', patchedData);
            } catch (error) {
                console.error('Error in patch request:', error);
            }
        };
        const GenAIResource = async (requestType) => {
            try {
                setObj(null)
                if (requestType === "create-genai") {
                    await fetch(Const.APIs.createGenAIAction(applicationName, applicationNamespace, destNamespace, resource.metadata.name), {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(requestType),
                    });
                }
                await new Promise(resolve => setTimeout(resolve, 800));
                await patchAnnotation();
                await new Promise(resolve => setTimeout(resolve, 800));
                const updateResponse = await fetch(Const.APIs.updateGenAIAction(applicationName, applicationNamespace, destNamespace, resource.metadata.name), {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify("update-genai")
                });
                if (!updateResponse.ok) throw new Error('Error updating GenAI action');

                await fetchStatus('');

            } catch (error) {
                console.error('Error initiating GenAI action:', error);
            }
        };

        const renderContent = () => {
            if (obj.status.phase === "running") {
                return <div><i className="fa fa-spinner" />
                    <div>
                        <img src="/images/intuit_assist_icon.svg" alt="thinking..." width="32" height="32"/> Thinking...
                    </div>
                </div>
            } else if (obj.status.phase === "failed") {
                return <div>Unable to fetch the response from GenAI. Retrying..</div>;
            } else if (obj.status.phase === "completed") {
                return (
                    <React.Fragment>
                        <div className="summary__header">
                            <h1>Summary</h1>
                        </div>
                        <button className={`argo-button argo-button--base`} onClick={async () => {
                            await GenAIResource("update-genai");
                        }}>
                            Summarize again
                        </button>
                        <div className="summary__header-info">
                            Last summarized at: &nbsp;
                            <Moment local={true} fromNow={true} ago={true}>
                                {obj.status.results[0].finishedAt}
                            </Moment> &nbsp; ago
                        </div>
                        <div className="summary__header-info">
                            Total time: &nbsp;
                            {formatTimeDifference(obj.spec.workflows[0].initiatedAt, obj.status.results[0].finishedAt)}
                        </div>
                        <div className="summary__feedback">
                            <FeedbackComponent />
                        </div>
                        <div className="summary__content">
                            <div className="summary__row">
                                <div className="summary__row__value">
                                    {obj.status.results[0].summary.mainSummary.split('-/-/-/-').map((part, index) => (
                                        <React.Fragment key={index}>
                                            {index === 0 && <div className='summary__row__value-text'>{part}</div>}
                                        </React.Fragment>
                                    ))}
                                </div>
                            </div>
                            <p className="warning">
                                <i className="fa fa-exclamation-triangle" /> Summary may display inaccurate info, including about custom resources, so double-check its responses
                            </p>
                            <div style={{ paddingTop: "50px", width: "450px" }} className="summary__row">
                                <ResourceCards app={application} />
                            </div>
                        </div>
                    </React.Fragment>
                );
            }
            return <div>
                No data available or fetch failed. &nbsp;
                <button className={`argo-button argo-button--base`} onClick={async () => {
                    await GenAIResource("update-genai");
                }}>Summarize again
                </button>
            </div>;
        };
        if (!isHealthy)  {
            return <div>App is healthy. No GenAI summary to display</div>;
        } else {
            return obj?.status ? renderContent() : <div>
                <img src="/images/intuit_assist_icon.svg" alt="thinking..." width="32" height="32"/>
                Thinking...
            </div>;            }
    };

