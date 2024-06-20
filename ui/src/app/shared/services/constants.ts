export const GenAI = 'gen-ai';
const API_VERSION = 'v1alpha1';
const SUPPORT_KIND = 'Support';
const SUPPORT_GROUP = 'argosupport.argoproj.extensions.io';
const ROLLOUT_KIND = 'Rollout';
const ROLLOUT_GROUP = 'argoproj.io';

export const APIs = {
    fetchGenAIStatus: (appName: string, appNamespace: string, destNamespace: string): string =>
        `/api/v1/applications/${appName}/resource?name=${GenAI}&appNamespace=${appNamespace}&resourceName=${GenAI}&namespace=${destNamespace}&version=${API_VERSION}&kind=${SUPPORT_KIND}&group=${SUPPORT_GROUP}`,
    patchAnnotation: (appName: string, appNamespace: string, destNamespace: string): string =>
        `/api/v1/applications/${appName}/resource?appNamespace=${appNamespace}&namespace=${destNamespace}&resourceName=${GenAI}&version=${API_VERSION}&kind=${SUPPORT_KIND}&group=${SUPPORT_GROUP}&patchType=${encodeURIComponent('application/merge-patch+json')}`,
    createGenAIAction: (appName: string, appNamespace: string, destNamespace: string, resourceName: string): string =>
        `/api/v1/applications/${appName}/resource/actions?appNamespace=${appNamespace}&namespace=${destNamespace}&resourceName=${resourceName}&version=${API_VERSION}&kind=${ROLLOUT_KIND}&group=${ROLLOUT_GROUP}`,
    updateGenAIAction: (appName: string, appNamespace: string, destNamespace: string, resourceName: string): string =>
        `/api/v1/applications/${appName}/resource/actions?appNamespace=${appNamespace}&namespace=${destNamespace}&resourceName=${GenAI}&version=${API_VERSION}&kind=${SUPPORT_KIND}&group=${SUPPORT_GROUP}`,
    getUserInfo: (): string => `/api/v1/session/userinfo`,
    getEvents: (appName: string, appNamespace: string, destNamespace: string, resourceName: string, resID: string): string => `/api/v1/applications/${appName}/events?appNamespace=${appNamespace}&resourceUID=${resID}&resourceNamespace=${destNamespace}&resourceName=${resourceName}`,
};

export const PHASE_RUNNING = "running";
export const PHASE_COMPLETED = "completed";
export const PHASE_ERROR = "error";

