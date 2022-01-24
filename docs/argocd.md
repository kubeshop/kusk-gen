# ArgoCD
![image](https://user-images.githubusercontent.com/14029650/129344427-1217687a-1d9c-490c-b877-d14d5e2638d9.png)

It is possible to make your OpenAPI schema the real source of truth for your K8s configurations with Kusk - by using it as an ArgoCD's [custom tool](https://argoproj.github.io/argo-cd/operator-manual/custom_tools)!

GitOps approach will allow you to use ArgoCD to automatically generate configurations in a determinative way based off your OpenAPI schema that you can store in Git to easily version, review and rollback your APIs. ArgoCD will take care of automatically syncing your configurations, while Kusk will happily generate it for you every time you make a change.

This guide will help you setup Kusk in ArgoCD.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Install Kusk as a custom tool](#install-kusk-as-a-custom-tool)
- [Use Kusk in ArgoCD UI](#use-kusk-in-argocd-ui)
- [Use Kusk in ArgoCD CLI](#use-kusk-in-argocd-cli)
- [Use Kusk in ArgoCD App manifest](#use-kusk-in-argocd-app-manifest)

## Prerequisites
1. Have a working ArgoCD installation - [ArgoCD quick start](https://argoproj.github.io/argo-cd/getting_started/).
2. Have a repository with OpenAPI schema configured within ArgoCD - for this example we will use Kusk's [example folder](./examples/petstore).

## Install Kusk as a custom tool
To install a custom tool in ArgoCD, we will follow the official [guide](https://argoproj.github.io/argo-cd/operator-manual/custom_tools/#adding-tools-via-volume-mounts).
We will add an `initContainer` to `argocd-repo-server` and a volume to download Kusk's binary to:
```diff
--- a/argocd-repo-server.yaml
+++ b/argocd-repo-server.yaml
@@ -31,6 +31,16 @@ spec:
                 topologyKey: kubernetes.io/hostname
               weight: 5
       automountServiceAccountToken: false
+      initContainers:
+        - name: download-kusk
+          image: alpine:3.8
+          command: [ sh, -c ]
+          args:
+            - wget -qO- https://github.com/kubeshop/kusk-gen/releases/download/0.0.1-rc1/kusk_0.0.1-rc1_Linux_x86_64.tar.gz | tar -xvzf - &&
+              mv kusk /custom-tools/
+          volumeMounts:
+            - mountPath: /custom-tools
+              name: kusk
       containers:
         - command:
             - uid_entrypoint.sh
@@ -72,7 +82,12 @@ spec:
               name: gpg-keyring
             - mountPath: /app/config/reposerver/tls
               name: argocd-repo-server-tls
+            - mountPath: /usr/local/bin/kusk
+              name: kusk
+              subPath: kusk
       volumes:
+        - name: kusk
+          emptyDir: {}
         - configMap:
             name: argocd-ssh-known-hosts-cm
           name: ssh-known-hosts
```

When installed, Kusk binary will be available during ArgoCD sync. To use that, we'll register a [configuration management plugin](https://argoproj.github.io/argo-cd/user-guide/application_sources/#config-management-plugins):
```diff
--- a/argocd-cm.yaml
+++ b/argocd-cm.yaml
@@ -5,3 +5,9 @@ metadata:
     app.kubernetes.io/name: argocd-cm
     app.kubernetes.io/part-of: argocd
   name: argocd-cm
+data:
+  configManagementPlugins: |
+    - name: kusk
+      generate:
+        command: ["/bin/sh", "-c"]
+        args: ["kusk $KUSK_GENERATOR -i $KUSK_INPUT $KUSK_ARGS"]
```

Once these changes are applied to your ArgoCD installation, you can begin to use Kusk!

## Use Kusk in ArgoCD UI
When you are [creating an App via UI](https://argoproj.github.io/argo-cd/getting_started/#creating-apps-via-ui), scroll down and select "Plugin" as a configuration management tool:
![image](https://user-images.githubusercontent.com/14029650/129340017-04ef2221-1793-4087-bf95-d60b1f2900d4.png)

After that, select Kusk as a plugin and fill the environment variables to specify the generator you want to invoke and the input file with your OpenAPI schema:
![image](https://user-images.githubusercontent.com/14029650/129340227-d729cb61-7c28-4869-80dd-9cea7153cfbd.png)

ArgoCD will sync your app and you will be able to see resources generated and applied to your cluster automatically 🪄:
![image](https://user-images.githubusercontent.com/14029650/129340502-e469fd2e-d745-483e-ba11-954dcc5c3ab2.png)

## Use Kusk in ArgoCD CLI
It is also possible to [create an App via CLI](https://argoproj.github.io/argo-cd/getting_started/#creating-apps-via-cli) with Kusk - just specify `--config-management-plugin kusk` option:
```shell
argocd app create petstore-kusk \
    --config-management-plugin kusk \
    --plugin-env KUSK_GENERATOR=ambassador \
    --plugin-env KUSK_INPUT=petstore_extension.yaml \
    --repo https://github.com/kubeshop/kusk-gen \
    --path examples/petstore \
    --dest-server https://kubernetes.default.svc \
    --dest-namespace default
```

## Use Kusk in ArgoCD App manifest
In ArgoCD it is possible to manage applications using App manifest and apply them using `kubectl`: [documentation](https://argoproj.github.io/argo-cd/operator-manual/declarative-setup/).
To use Kusk in this setup, add Kusk in a `plugin` node:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: petstore-kusk
spec:
  destination:
    namespace: default
    server: 'https://kubernetes.default.svc'
  source:
    path: examples/petstore
    repoURL: 'https://github.com/kubeshop/kusk-gen'
    targetRevision: main
    plugin:
      name: kusk
      env:
        - name: KUSK_INPUT
          value: petstore_extension.yaml
        - name: KUSK_GENERATOR
          value: ambassador
  project: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```
