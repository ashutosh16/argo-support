import * as models from "argocd/ui/src/app/shared/models"
import * as classNames from 'classnames';
import * as React from 'react';


export const ARGO_SUCCESS_COLOR = '#18BE94';
export const ARGO_WARNING_COLOR = '#f4c030';
export const ARGO_FAILED_COLOR = '#E96D76';
export const ARGO_RUNNING_COLOR = '#0DADEA';
export const ARGO_GRAY4_COLOR = '#CCD6DD';
export const ARGO_GRAY6_COLOR = '#6D7F8B';
export const ARGO_TERMINATING_COLOR = '#DE303D';
export const ARGO_SUSPENDED_COLOR = '#766f94';


//Copy the code from argocd
export const COLORS = {
    connection_status: {
        failed: ARGO_FAILED_COLOR,
        successful: ARGO_SUCCESS_COLOR,
        unknown: ARGO_GRAY4_COLOR
    },
    health: {
        degraded: ARGO_FAILED_COLOR,
        healthy: ARGO_SUCCESS_COLOR,
        missing: ARGO_WARNING_COLOR,
        progressing: ARGO_RUNNING_COLOR,
        suspended: ARGO_SUSPENDED_COLOR,
        unknown: ARGO_GRAY4_COLOR
    },
    operation: {
        error: ARGO_FAILED_COLOR,
        failed: ARGO_FAILED_COLOR,
        running: ARGO_RUNNING_COLOR,
        success: ARGO_SUCCESS_COLOR,
        terminating: ARGO_TERMINATING_COLOR
    },
    sync: {
        synced: ARGO_SUCCESS_COLOR,
        out_of_sync: ARGO_WARNING_COLOR,
        unknown: ARGO_GRAY4_COLOR
    },
    sync_result: {
        failed: ARGO_FAILED_COLOR,
        synced: ARGO_SUCCESS_COLOR,
        pruned: ARGO_GRAY4_COLOR,
        unknown: ARGO_GRAY4_COLOR
    },
    sync_window: {
        deny: ARGO_FAILED_COLOR,
        allow: ARGO_SUCCESS_COLOR,
        manual: ARGO_WARNING_COLOR,
        inactive: ARGO_GRAY4_COLOR,
        unknown: ARGO_GRAY4_COLOR
    }
};


export const HealthStatusIcon = ({state, noSpin}: {state: models.HealthStatus; noSpin?: boolean}) => {
    let color = COLORS.health.unknown;
    let icon = 'fa-question-circle';

    switch (status) {
        case models.HealthStatuses.Healthy:
            color = COLORS.health.healthy;
            icon = 'fa-heart';
            break;
        case models.HealthStatuses.Suspended:
            color = COLORS.health.suspended;
            icon = 'fa-pause-circle';
            break;
        case models.HealthStatuses.Degraded:
            color = COLORS.health.degraded;
            icon = 'fa-heart-broken';
            break;
        case models.HealthStatuses.Progressing:
            color = COLORS.health.progressing;
            icon = `fa fa-circle-notch ${noSpin ? '' : 'fa-spin'}`;
            break;
        case models.HealthStatuses.Missing:
            color = COLORS.health.missing;
            icon = 'fa-ghost';
            break;
    }

    return React.createElement('i', {
        'qe-id': 'utils-health-status-title',
        title: status,
        className: classNames('fa', icon),
        style: {color}
    });
};


export const ComparisonStatusIcon = ({
                                         status,
                                         resource,
                                         label,
                                         noSpin
                                     }: {
    status: models.SyncStatusCode;
    resource?: {requiresPruning?: boolean};
    label?: boolean;
    noSpin?: boolean;
}) => {
    let className = 'fas fa-question-circle';
    let color = COLORS.sync.unknown;
    let title: string = 'Unknown';

    switch (status) {
        case models.SyncStatuses.Synced:
            className = 'fa fa-check-circle';
            color = COLORS.sync.synced;
            title = 'Synced';
            break;
        case models.SyncStatuses.OutOfSync:
            // eslint-disable-next-line no-case-declarations
            const requiresPruning = resource && resource.requiresPruning;
            className = requiresPruning ? 'fa fa-trash' : 'fa fa-arrow-alt-circle-up';
            title = 'OutOfSync';
            if (requiresPruning) {
                title = `${title} (This resource is not present in the application's source. It will be deleted from Kubernetes if the prune option is enabled during sync.)`;
            }
            color = COLORS.sync.out_of_sync;
            break;
        case models.SyncStatuses.Unknown:
            className = `fa fa-circle-notch ${noSpin ? '' : 'fa-spin'}`;
            break;
    }
    return React.createElement('i', { 'key': 'sync-status-icon',
        'qe-id': 'utils-sync-status-title',
        title: status,
        className: className,
        style: {color}
    });
}