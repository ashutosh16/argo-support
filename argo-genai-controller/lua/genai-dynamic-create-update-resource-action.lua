  resource.customizations.actions.argoproj.io_Rollout: |
    discovery.lua: |
       actions = {}
       actions["create-genai"] = {}
       return actions
    definitions:
    - name: create-genai
      action.lua: |
        local os = require("os")
        local genaiObj = {}
        local spec = {}
        local ownerRef = {}

        genaiObj.apiVersion = "argosupport.argoproj.extensions.io/v1alpha1"
        genaiObj.kind = "Support"
        genaiObj.metadata = {}
        genaiObj.metadata.name = "gen-ai"
        genaiObj.metadata.namespace = obj.metadata.namespace
        genaiObj.metadata.labels = {}
        genaiObj.metadata.labels["app.kubernetes.io/instance"] = obj.metadata.labels["app.kubernetes.io/instance"]
        genaiObj.metadata.labels["rollouts-pod-template-hash"] = obj.status.currentPodHash
        genaiObj.metadata.annotations = {}
        genaiObj.metadata.annotations["rollout.argoproj.io/revision"] = obj.metadata.annotations["rollout.argoproj.io/revision"]

        ownerRef.apiVersion = obj.apiVersion
        ownerRef.kind = obj.kind
        ownerRef.name = obj.metadata.name
        ownerRef.uid = obj.metadata.uid
        ownerRef.blockOwnerDeletion = true
        ownerRef.controller = true
        genaiObj.metadata.ownerReferences = {}
        genaiObj.metadata.ownerReferences[1] = ownerRef

        local workflows = {}
        workflows.name = "gen-ai"
        local datetime = os.date("!%Y-%m-%dT%H:%M:%SZ")
        workflows.initiatedAt = datetime
        workflows.configMapRef = {}
        workflows.configMapRef.name =  "genai-cm"
        workflows.autProviderRef = {}
        workflows.autProviderRef[1] = {}
        workflows.autProviderRef[1].name = "genai-authprovider"
        workflows.autProviderRef[2] = {}
        workflows.autProviderRef[2].name = "argocd-auth-provider"
        spec.workflows = {}
        spec.workflows[1] = workflows
        genaiObj.spec = spec
        impactedResource = {}
        impactedResource.operation = "create"
        impactedResource.resource = genaiObj
        local result = {}
        result[1] = impactedResource
        return result

  resource.customizations.actions.argosupport.argoproj.extensions.io_Support: |
    discovery.lua: |
       actions = {}
       actions["update-genai"] = {}
       return actions
    definitions:
    - name: update-genai
      action.lua: |
        local os = require("os")
         if not obj or not obj.spec or not obj.spec.workflows then
          error("Object is  missing required fields")
        end
        local datetime = os.date("!%Y-%m-%dT%H:%M:%SZ")
        obj.spec.workflows[1].initiatedAt = datetime
        impactedResource = {}
        impactedResource.operation = "patch"
        impactedResource.resource = obj
        local result = {}
        result[1] = impactedResource
        return result