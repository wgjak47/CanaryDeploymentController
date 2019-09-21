# CanaryDeploymentController
An sample project for learning kubebuilder. 
This controller aims to manage similar kubernetes deployments

## Build
```shell
# note don't run make or make manifests
# this will generate wrong validations
make manager
```

## Run
```shell
# create crd
kubectl create -f config/crd/webapp.wgjak47.me_canarydeployments.yaml
# create sample yaml
kubectl apply -f config/sample/webapp_v1_canarydeployments.yaml
bin/manager
```

## Example
```yaml
apiVersion: webapp.wgjak47.me/v1
kind: CanaryDeployment
metadata:
  name: canarydeployment-sample
spec:
  # the common spec of deployments
  commonSpec:
    replicas: 1 # tells deployment to run 2 pods matching the template
    selector:
      matchLabels:
        app: nginx
    template: # create pods using pod definition in this template
      metadata:
        # unlike pod-nginx.yaml, the name is not included in the meta data as a unique name is
        # generated from the deployment name
        labels:
          app: nginx
      spec:
        containers:
        - name: nginx
          image: nginx:1.7.9
          ports:
          - containerPort: 80
  appSpecs:
    # define your deployment with patch on commonSpec
    # this key below will be a part of name of deployment
    newer:
      # support json merge patch or k8s strategic merge
      # use 'merge' for json merge patch
      # 'strategic' for k8s strategic merge
      Type: merge
      # right now you have to use json string here
      # I have encounter some problem on set this field 
      # as map[string]interface{}
      Spec: '{
        "template": {
          "spec": {
            "containers": [{
             "name": "nginx",
              "image": "nginx:latest"
              }
            ]
          }
        }
      }'
    older:
      Type: merge
      Spec: '{}'
```

## Roadmap
- [ ] suite test
- [ ] webhook for validations
- [ ] support JSON Patch (rfc6902)[https://tools.ietf.org/html/rfc6902]
