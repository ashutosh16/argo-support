import React, { useEffect, useState, useCallback } from 'react';
import { Tabs, Space } from 'antd';
import classNames from 'classnames';
import '../../styles.scss';
import { Summary } from "../summary/summary";
import {MetricsWrapper} from "../metrics/metricswrapper";
import { Rollouts } from "../rollouts";

const cx = classNames;

interface Props {
    application: any;
    tree: any;
}

const Label = ({ label }) => (
    <Space size='small'>
        <span className={cx('metric-label')} title={label}>
            {label}
        </span>
    </Space>
);

export const Panel = ({ application, tree }: Props) => {
    const [activeKey, setActiveKey] = useState('genai-summary');
    const [resNode, setResNode] = useState(null);
    const [isLoading, setIsLoading] = useState(true); // Added loading state
    const application_name = application?.metadata?.name || "";
    useEffect(() => {
        const rollouts = application.status.resources.filter(resource => resource.kind === "Rollout");
        const rObj = rollouts.length > 0 ? rollouts[0] : null;
        if (!rObj) {
            setIsLoading(false);
            return;
        }

        const url = `/api/v1/applications/${application_name}/resource?name=${rObj.name}&appNamespace=${application.metadata.namespace}&namespace=${application.spec.destination.namespace}&resourceName=${rObj.name}&version=v1alpha1&kind=Rollout&group=argoproj.io`;

        fetch(url)
            .then(response => response.json())
            .then(data => {
                setResNode(typeof data.manifest === "string" ? JSON.parse(data.manifest) : data.manifest);
                setIsLoading(false);
            })
            .catch(err => {
                console.error("Error fetching rollout data:", err);
                setIsLoading(false);
            });

    }, [application_name, application.metadata.namespace, application.spec.destination.namespace]);

    const RenderTabContent = useCallback(({ key }) => {
        switch (key) {
            case 'genai-summary':
                return <Summary application={application} tree={tree}  resource={resNode} />;
            case 'rollout':
                return <Rollouts application={application} resource={resNode} tree={tree} IsDisplay={true}/>;
            case 'metrics':
                return <MetricsWrapper application={application} resource={resNode} tree={tree} IsDisplay={true}/>;
            default:
                return null;
        }
    }, [application, resNode, tree]);

    const tabItems = [
        {
            label: <Label label='Intuit Assist' />,
            key: 'genai-summary',
            children: activeKey === 'genai-summary' ? RenderTabContent({ key: 'genai-summary' }) : null
        },
        {
            label: <Label label='Rollout' />,
            key: 'rollout',
            children: activeKey === 'rollout' ? RenderTabContent({ key: 'rollout' }) : null
        },
        {
            label: <Label label='Metrics' />,
            key: 'metrics',
            children: activeKey === 'metrics' ? RenderTabContent({ key: 'metrics' }) : null
        },
    ];

    const onTabChange = (key) => {
        setActiveKey(key);
    };

    if (isLoading) {
        return <div>Loading...</div>;
    }


    return (
        <Tabs
            className={cx('tabs')}
            items={tabItems}
            activeKey={activeKey}
            onChange={onTabChange}
            tabPosition='top'
            size='small'
            tabBarGutter={12}
        />
    );
};