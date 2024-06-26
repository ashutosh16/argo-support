import React, { useState, useEffect } from 'react';
import './summary.scss';
import { FeedbackComponent } from '../feedback/feedback';
import { ResourceCards } from '../resourcecard/resourcecard';
import {getGenResApi, actionsApi, patchApi} from '../../shared/services/client';
import Moment from 'react-moment';
import * as models from "argocd/ui/src/app/shared/models";
import ReactMarkdown from 'react-markdown'

const RetryTimer = 300000;

interface SummarysProps{
    application: models.Application;
    tree: models.ApplicationTree;
    resource: any
}

function IsEligibleForRequest(genaiObject ){
    if (genaiObject == null) { return true; }
    const phase = genaiObject?.status?.phase
    const time = genaiObject && genaiObject?.status.lastTransitionTime
    const timeLapsed =  time && new Date().getTime() - new Date(time).getTime() > RetryTimer;
    return phase === "error" || timeLapsed
}
export const Summary = (summaryProps: SummarysProps) => {
    const {application: application, tree, resource} = summaryProps;
    const [genObj, setGenObj] = useState(null);
    const [isLoading, setIsLoading] = useState(false);

    const isHealthy = application.status?.health?.status !== "Healthy" || application.status?.conditions;

    useEffect(() => {
        const fetchGenAIData = async () => {

            let resData = await getGenResApi(application, resource);
            if (!isLoading && genObj == null) {
                setIsLoading(true);
                if (resData == null) {
                    await actionsApi("create-genai", application, resource);
                } else {
                    if (IsEligibleForRequest(resData)) {
                        await actionsApi("update-genai", application, resource);
                    }
                }
            }
            resData = await getGenResApi(application, resource);
            setGenObj(resData);
            setIsLoading(false);

        };
        if (isHealthy) {
            fetchGenAIData();
        }
    }, []);
        const result = genObj &&  (genObj?.status?.results || [] ).length > 0  && genObj?.status?.results[0];
        const phase = genObj?.status && genObj.status?.phase;
        const degradedResource = application.status.resources.filter(res => res.health?.status  !== models.HealthStatuses.Healthy).length
        //const degradedResourcesCount = application && application.status.resources.filter(res => res.health?.status != models.HealthStatuses.Healthy).length;
    return (
            !isHealthy?
             <div>App is healthy. No GenAI summary to display</div>:
               <React.Fragment>
                <div className="summary">
                    <span className="summary__health-title">App Health: {application.status?.health?.status}</span>
                    <hr className="summary__health-divider" />

                    <div className="summary__container" key={result?.name}>
                    {  phase!== "completed" ?
                        phase === "failed" ?
                            <div className="summary__container__main">Retrying..</div>
                            : phase === "error" ?
                                <div className="summary__container__main">Failed to fetch the summary, try after sometime.
                                </div>
                                :<div className="summary__container__main">
                                <svg width="24" height="24" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                                    <path fillRule="evenodd" clipRule="evenodd" d="M8.41096 1.65967L8.75589 4.94102C8.85251 5.87097 7.90762 6.56062 7.051 6.17895L4.03737 4.83697L4.53103 3.72838L7.54466 5.07036L7.5462 5.07101L7.54736 5.07024L7.54902 5.06885L7.54886 5.06661L7.20407 1.78653L8.41096 1.65967ZM11.2679 3.5228L10.061 3.39601L9.7101 6.7363C9.61222 7.67514 10.5662 8.37002 11.4301 7.98605L14.4983 6.6198L14.0047 5.5112L10.9367 6.87732C10.9325 6.87918 10.9303 6.87949 10.9301 6.87951C10.9298 6.87946 10.9268 6.8789 10.9228 6.87599C10.9188 6.87308 10.9172 6.87003 10.917 6.86957C10.9169 6.86937 10.9166 6.8672 10.917 6.86261L11.2679 3.5228ZM10.0204 9.42133L13.3038 8.72359L13.5561 9.91062L10.2724 10.6084C10.2701 10.6089 10.2684 10.6094 10.2676 10.6097L10.2663 10.6103C10.2659 10.6107 10.264 10.6129 10.2626 10.6173C10.2612 10.6215 10.2614 10.6243 10.2615 10.6251L10.2621 10.626L10.2634 10.6278L10.265 10.6297L10.2656 10.6303L12.5116 13.1252L11.6097 13.9371L9.36425 11.4429C8.73182 10.7419 9.09616 9.61755 10.0204 9.42133ZM5.43368 10.5505C5.90592 9.73236 7.08756 9.73241 7.55971 10.5507L9.23836 13.458L8.18741 14.0648L6.50868 11.1573C6.50748 11.1553 6.50647 11.1538 6.50593 11.1531L6.50501 11.152C6.50467 11.1518 6.5018 11.1505 6.49665 11.1505C6.49165 11.1505 6.4888 11.1518 6.48833 11.152L6.48738 11.1531C6.48683 11.1538 6.48591 11.1551 6.48471 11.1572L4.80589 14.0648L3.75495 13.458L5.43368 10.5505ZM5.03881 6.54493L1.823 5.8068L1.55151 6.98959L4.76607 7.72743L4.76824 7.72801L4.76902 7.73003L4.76937 7.73138L4.76825 7.73262L2.52128 10.148L3.40979 10.9746L5.65677 8.5592C6.29566 7.8727 5.94981 6.7552 5.03881 6.54493Z" fill="#236CFF"/>
                                </svg> Thinking...
                            </div>
                        :

                        <div className="summary__container__main">
                            <div className="summary__container__main__analysis-title">Analysis
                                <span className="last-summarized-timestamp">

                                     Last summarized at: &nbsp;
                                    <Moment local={true} fromNow={true} ago={true}>
                                             {result && result.finishedAt}
                                         </Moment>  ago
                                 </span>
                            </div>

                            <div className="content">
                                <ReactMarkdown children={result.summary.mainSummary} />
                                <span style={{ fontSize: '14px', fontWeight: 100 }}>
                            Powered by GenAI - results may be inaccurate and not reflect Argo's view
                            </span>
                            </div>


                            {result &&  <FeedbackComponent patchApi={ (patch) => {
                                patchApi(application, resource, patch)
                                actionsApi("update-genai", application, resource);

                            }} result={result}/>}

                             </div>
                    }
                        <div className="summary__container__feedback">
                            {result?.help && (
                                <div className="summary__container__feedback-help">
                                    Sources
                                    <div className="summary__container__feedback-help-link">
                                        <span>Slack Channel: <a href={`https://slack.com/app_redirect?channel=${result?.help.slackChannel}`} target="_blank" rel="noopener noreferrer">
                                            {result?.help.slackChannel}</a></span>
                                        Links:
                                            {(result?.help.links || []).map((link, index) => {
                                                const url = new URL(link);
                                                return (
                                                        <a  key= {index} href={url.href} target="_blank" rel="noopener noreferrer">
                                                            {url.host}
                                                        </a>
                                                );
                                            })}
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>

                </div>
                   {
                       degradedResource > 0 && <div className="summary__resourcecard">
                       <label className="summary__resourcecard-title">
                           Unhealthy Resources
                       </label>

                       <div className="summary__resourcecard-container">
                           <ResourceCards  app={application} tree={tree}/>
                       </div>
                   </div>
                   }
               </React.Fragment>
        );
};

