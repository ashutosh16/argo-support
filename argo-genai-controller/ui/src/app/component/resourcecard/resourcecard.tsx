import React, {useEffect, useState} from 'react';
import * as models from "argocd/ui/src/app/shared/models"
import './resourcecard.scss';
import {ComparisonStatusIcon, HealthStatusIcon} from "../../shared/utils";
import {ResourceIcon} from "argocd/ui/src/app/applications/components/resource-icon";
import {ResourceLabel} from 'argocd/ui/src/app/applications/components/resource-label';
import {getEvents} from '../../shared/services/client';

interface ResourceCardsProps{
app: models.Application;
tree: models.ApplicationTree;
}


export const ResourceCards = (
    props: ResourceCardsProps
) => {

    const resNode: models.ResourceStatus[] = props.app && props.app.status.resources.filter(res => res.health?.status && res.health?.status  !== models.HealthStatuses.Healthy)
        .sort((a, b) => a.health?.status.localeCompare(b.health?.status))
    const [events, setEvents] = useState<{ [key: string]: any[] }>({});

    useEffect(() => {
        const fetchEvents = async () => {
            const eventsMap: { [key: string]: any[] } = {};
            for (const res of props.app.status.resources) {
                try {
                    const node: models.ResourceNode = props.tree.nodes.find((node) => node.name === res.name && node.kind === res.kind && node.namespace === res.namespace);
                    if (res != null && node?.uid !== undefined) {
                        eventsMap[res.name] = await getEvents(props.app, res, node?.uid);
                    }
                } catch (error) {
                    console.error('Error fetching events:', error);
                }
            }
            props.app && setEvents(eventsMap);
        };

        fetchEvents();
    }, []);

    return (
        <>
            {resNode.map(res => (
                <div key ={res.kind+res.namespace+res.name}
                    className='applications-tiles argo-table-list argo-table-list--clickable row small-up-1 medium-up-2 large-up-3 xxxlarge-up-4' style={{margin: '2px'}}>
                    <div
                        className={`argo-table-list__row applications-list__entry applications-list__entry--health-${res.health?.status}`} style={{width:'350px'}}>
                        <div
                            className={`columns small-12 applications-list__info  applications-tiles__item`} >
                            <div className='row' style={{display: "inline-block", paddingBottom: '20px', paddingLeft: '50px'}}>
                                <div className='application-resource-tree__node-kind-icon'>
                                    <ResourceIcon kind={res.kind} />
                                    <br />
                                    <div className='application-resource-tree__node-kind' style={{display: "flex", justifyContent: "center",
                                        alignItems: "flex-end"}}>
                                        {ResourceLabel({kind: res.kind})}</div>
                                </div>  &nbsp;
                                 <span style={{fontSize: '16px'}}> {res?.name}</span>
                            </div>
                            <div className='row'>
                                    HEALTH STATUS: &nbsp;
                                { res?.status && <div  qe-id='applications-tiles-health-status'>

                                        <React.Fragment>
                                            <HealthStatusIcon state={res?.health} /> {res?.health?.status} &nbsp;
                                        </React.Fragment>
                                      &nbsp;

                                    {res && <ComparisonStatusIcon status={res?.status} resource={res} label={true} />}
                                </div>
                                }
                            </div>

                            <div className='row'>
                                EVENTS:  &nbsp;
                                    {events[res.name] && events[res.name].filter(event => event.type === 'warning' || event.type === 'error').length > 0 ? (
                                        events[res.name].filter(event => event.type === 'warning' || event.type === 'error').map((event, index) => (
                                            <div key={index} className="resource__item-value">
                                                {event.message}
                                            </div>
                                        ))
                                    ) : (
                                         ' No active events'
                                    )}
                            </div>
                            <div className='row'>
                                ERROR MESSAGE: &nbsp; <br />
                                {res?.health?.message ? (
                                    <div className='resource__item-value' style={{backgroundColor: '#eff3f5db'}}>
                                        {res.health.message}
                                    </div>
                                ) : (
                                    'No error message'
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            ))}
        </>
    );
}