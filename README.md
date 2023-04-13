# dapr-fix-failed-injection

This is a simple program to fix the failed injection of Dapr sidecar by killing the pod and let the Kubernetes restart it.

## Usage

create below CronJob to run the program every 5 minutes

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: fix-failed-dapr-injection-job
  namespace: dapr-system
spec:
  schedule: "*/5 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: fix-failed-dapr-injection
              image: ghcr.io/heavenwing/dapr-fix-failed-injection:main
          serviceAccountName: dapr-operator
          restartPolicy: "Never"
```

and deploy it into your Kubernetes cluster.

NOTE: it will check default namespace only, if you want to check other namespaces, you can pass -ns=OTHER_NAMESPACE flag to the program.