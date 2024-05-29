import React, { useState, useEffect } from 'react';
import './summary.scss';
import { FeedbackComponent } from '../feedback/feedback';
import { ResourceCards } from '../resourcecard/resourcecard';

const GenAI = 'gen-ai';

const apiUrls = {
    fetchGenAIStatus: (appName, appNamespace, destNamespace) =>
        `/api/v1/applications/${appName}/resource?name=${GenAI}&appNamespace=${appNamespace}&resourceName=${GenAI}&namespace=${destNamespace}&version=v1alpha1&kind=Support&group=argosupport.argoproj.extensions.io`,
    createGenAIAction: (appName, appNamespace, destNamespace, resourceName) =>
        `/api/v1/applications/${appName}/resource/actions?appNamespace=${appNamespace}&namespace=${destNamespace}&resourceName=${resourceName}&version=v1alpha1&kind=Rollout&group=argoproj.io`,
    updateGenAIAction: (appName, appNamespace, destNamespace, resourceName) =>
        `/api/v1/applications/${appName}/resource/actions?appNamespace=${appNamespace}&namespace=${destNamespace}&resourceName=${GenAI}&version=v1alpha1&kind=Support&group=argosupport.argoproj.extensions.io`
,
};

const isFetchNeeded = (lastTransitionTime) => {
    const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000).getTime();


    return lastTransitionTime > fiveMinutesAgo;
};

export const Summary = ({ application, resource }) => {
    const [result, setResult] = useState(null);
    const [isLoading, setIsLoading] = useState(true);
    const [buttonDisabled, setButtonDisabled] = useState(true);


    const { metadata, spec } = application;
    const applicationName = metadata?.name || '';
    const applicationNamespace = metadata?.namespace || '';
    const destNamespace = spec.destination.namespace || '';



    useEffect(() => {
        if (application?.status?.health?.status !== "Healthy" || (application?.status?.conditions && application.status.conditions.length > 0)) {
            const actions = async () => {
                try {
                    await initiateAction("create-genai");
                    await fetchGenAIResource();
                    const interval = setInterval(fetchGenAIResource, 20000);
                    return () => {
                        setIsLoading(false)
                        clearInterval(interval);
                    };
                } catch (error) {
                    console.error("Error in setting up GenAI actions:", error);
                }
            };
            actions();
        }
    }, []);

    const fetchGenAIResource = async () => {

        const fetchData = async () => {

            try {
            const response = await fetch(apiUrls.fetchGenAIStatus(applicationName, applicationNamespace, destNamespace));
            if (!response.ok) initiateAction("create-genai")
            const data = await response.json();

                const jsonData = typeof data.manifest === "string" ? JSON.parse(data.manifest) : data.manifest;
                setResult(jsonData);

                    const lastTransitionTime = new Date(jsonData?.status?.lastTransitionTime).getTime();
                    setButtonDisabled(isFetchNeeded(lastTransitionTime));
                    setIsLoading(false);


            } catch (error) {
                setIsLoading(false);
            }
        };
        fetchData();
    };

    const initiateAction = async (requestType: string) => {
        setIsLoading(true);

        try {
            let url;
            if (requestType === "create-genai") {
                url = apiUrls.createGenAIAction(applicationName, applicationNamespace, destNamespace, resource.metadata.name);
            } else {
                url = apiUrls.updateGenAIAction(applicationName, applicationNamespace, destNamespace, resource.metadata.name);
            }
            const response = await fetch(url, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(requestType),
            });

            if (response.ok) {
                fetchGenAIResource();
            } else if (response.status === 409) {
                console.log('GenAI resource already exists in the namespace');
            }
        } catch (error) {
            console.error('Error initiating GenAI action:', error);
        }
    };


    const renderContent = () => {
        if (result.status.phase === "running") {
            return <div><i className="fa fa-spinner" /> GenAI process is running. Wait...</div>;
        } else if (result.status.phase === "failed") {
            return <div>Unable to fetch the response from GenAI.Retry..</div>;
        } else if (result.status.phase === "completed") {
            return (
                <React.Fragment>
                    <div className="summary__header">
                        <h1>Summary</h1>
                    </div>
                    <button disabled={buttonDisabled} className={`${buttonDisabled ? 'summary__btn-disable' : ''} argo-button argo-button--base`} onClick={() =>  initiateAction("update-genai")}>
                        Summarize again
                    </button>
                    {buttonDisabled && <small style={{ marginLeft: '10px' }}>Wait for 5 mins</small>}
                    <div className="summary__feedback">
                        <FeedbackComponent />
                    </div>
                    <div className="summary__content">
                        <div className="summary__row">
                            <div className="summary__row__label">Overview</div>
                            <div className="summary__row__value">
                                {result.status.results[0].summary.mainSummary.split('-/-/-/-').map((part, index) => (
                                    <React.Fragment key={index}>
                                        {index === 0 && <p>{part}</p>}
                                    </React.Fragment>
                                ))}
                            </div>
                        </div>
                        <p className="warning">
                            <i className="fa fa-exclamation-triangle" /> Summary may display inaccurate info, including about deployment, so double-check its responses
                        </p>
                        <div style={{ paddingTop: "50px" }} className="summary__row">
                            <ResourceCards app={application} />
                        </div>
                    </div>
                </React.Fragment>
            );
        }
        return <div>
            No data available or fetch failed. &nbsp;
            <button className={`argo-button argo-button--base`} onClick={() =>  initiateAction("update-genai")}>
                Summarize again
            </button>
        </div>;
    };

    if (application?.Status?.Conditions > 0)  {
        return <div>App is healthy. No GenAI summary to display</div>;
    } else {
        return !isLoading && result?.status ? renderContent() : <div><i className="fa fa-spinner" /> Fetching data...</div>;
    }
};

