# default-imagetag-webhook
This webhook is implemented using the kubewebhook library.  
If you deploy a pod without tagging .spec.containers[].image, it will be given the latest tag (I understand that kubernetes considers it up-to-date even if it is not given the latest tag; I wanted to change the look of the kubectl describe).
```golang
func imageTagMutator(_ context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return &kwhmutating.MutatorResult{}, nil
	}

	stopLen := 1
	for i, c := range pod.Spec.Containers {
		s := strings.Split(c.Image, ":")
		if len(s) == stopLen {
			pod.Spec.Containers[i].Image = fmt.Sprintf("%s%s", c.Image, ":latest")
		}
	}

	return &kwhmutating.MutatorResult{
		MutatedObject: pod,
	}, nil
}
```

# Build
```
$ export IMG=default-imagetag-webhook:latest
$ docker build . -t ${IMG}
```

# Deploy
Issue certificates with [cert-manager](https://cert-manager.io/docs/) and inject to Pod and AdminssionWebhook. Therefore, deploy cert-manager first.
```
$ kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.2/cert-manager.yaml
$ kubectl apply -f deploy/webhook-issuer.yaml
$ kubectl apply -f deploy/webhook.yaml
$ kubectl apply -f deploy/webhook-registration.yaml
```

# Usage
After deployment is complete, you can check the log.  
If you deploy a pod without adding a tag to .spec.containers[].image, the following log will be output and you can see what operations were performed.
```
$ kubectl logs default-imagetag-webhook-5bd6b8b685-t26t4 -f
time="2022-07-11T00:41:14Z" level=info msg="Listening on :8080"
time="2022-07-11T00:42:12Z" level=debug msg="Webhook mutating review finished with: '[{\"op\":\"replace\",\"path\":\"/spec/containers/0/image\",\"value\":\"nginx:latest\"}]' JSON Patch" dry-run=false kind=v1/Pod name= ns=default op=create path=/mutate request-id=682f8ab2-6ace-4663-a748-96299cf4f650 trace-id= webhhok-type=mutating webhook-id=imageTagMutator webhook-kind=mutating wh-version=v1
time="2022-07-11T00:42:12Z" level=info msg="Admission review request handled" dry-run=false duration=276.777541ms kind=v1/Pod name= ns=default op=create path=/mutate request-id=682f8ab2-6ace-4663-a748-96299cf4f650 svc=http.Handler trace-id= webhook-id=imageTagMutator webhook-kind=mutating wh-version=v1
```
