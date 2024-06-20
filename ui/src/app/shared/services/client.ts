import * as Const from "./constants";


const controller = new AbortController();
const signal = controller.signal;



export const getGenResApi = async (application: any, resource: any) => {
    const applicationName = application.metadata?.name || '';
    const applicationNamespace = application.metadata?.namespace || '';
    const destNamespace = application.spec?.destination?.namespace || '';
    try {
        let jsonData = null;
        do {
            const statusResponse = await fetch(Const.APIs.fetchGenAIStatus(applicationName, applicationNamespace, destNamespace), { signal });
            if (statusResponse.ok) {
                const data = await statusResponse.json();
                jsonData = typeof data.manifest === "string" ? JSON.parse(data.manifest) : data.manifest;
            } else {

                break;
            }
            if (jsonData?.status?.phase === "completed" || jsonData?.status?.phase === "error") {
                break;
            }
            //loop until the status is completed or error
            await new Promise(resolve => setTimeout(resolve, 5000));
        } while (true);
        return jsonData;
    } catch (error) {
        controller.abort();
        console.error('Error fetching GenAI status:', error);
    }
};

export const actionsApi = async (requestType, application: any, resource: any) => {
    const applicationName = application.metadata?.name || '';
    const applicationNamespace = application.metadata?.namespace || '';
    const destNamespace = application.spec?.destination?.namespace || '';
    const resName = resource?.metadata.name || '';
    console.log("Action GenAI data");

    if (requestType === "create-genai") {
        await fetch(Const.APIs.createGenAIAction(applicationName, applicationNamespace, destNamespace, resName), {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(requestType),
        });
    }

    const annotation = JSON.stringify({
        "metadata": {
            "annotations": {
                "argosupport.argoproj.extensions.io/argo-app": `${JSON.stringify(application.status)}`
            }
        }
    });
    await patchApi(application, resource, annotation);
    const updateResponse = await fetch(Const.APIs.updateGenAIAction(applicationName, applicationNamespace, destNamespace, resource.metadata.name), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify("update-genai")
    });
    if (!updateResponse.ok) throw new Error('Error updating GenAI action');
};


export const patchApi = async (application: any,resource: any,  patch: string) => {
    const applicationName = application.metadata?.name || '';
    const applicationNamespace = application.metadata?.namespace || '';
    const destNamespace = application.spec?.destination?.namespace || '';
    const url = Const.APIs.patchAnnotation(applicationName, applicationNamespace, destNamespace);

    try {
        const response = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(patch)
        });

        if (!response.ok) throw new Error('error patching the genai resource');
    } catch (error) {
        console.error('Error in patch request:', error);
    }
};

export const getEvents = async (application: any, resource: any, resID: string) => {
    const applicationName = application.metadata?.name || '';
    const applicationNamespace = application.metadata?.namespace || '';
    const destNamespace = application.spec?.destination?.namespace || '';
    const resName = resource.name || '';
    const url = Const.APIs.getEvents(applicationName, applicationNamespace, destNamespace, resName, resID);

    try {
        const response = await fetch(url, {
            method: 'GET',
            headers: { 'Content-Type': 'application/json' }
        });
        if (response.ok) {
            const data = await response.json();
            return data.items;
        } else {
            throw new Error('Error fetching events');
        }
    } catch (error) {
        console.error('Error in getEvents request:', error);
    }
};

