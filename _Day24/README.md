從異世界歸來的第二四天 - Kubernetes Resources(四) - Metrics Server 安裝
---

## 概述

在前面幾天中我們學習了許多關於資源設定的觀念，但漸漸會開始發現一個事，我們該怎麼知道以及監控所有服務的資源利用率以及健康狀況等等， `Kubernetes` 有許多指標數據需要收集，大致紹可以分為集群本身以及 `Pod` ，包含節點是否正常運行，像是 Disk, CPU, Memory 利用率和需要跟 `Kubernetes` 獲取的佈署副本數、存活監測、健康監測，都需要一個工具來替我們收集資源指標並且整合，而這個工具 `Kubernetes` 本身並沒有提供，需要我們使用到擴充套件來實現。

### Metrics-Server

`Docker-Desktop` 提供的 `Kubernetes` 沒有幫我們預設安裝 Metrics Server，而像是 GKE 在一開始使用上就可以使用 Google 自帶的資源監控服務，這些資源指標 API 透過 API Server `/apis/metrics.k8s.io` 進行存取，Metrics Server 是集群級別的數據聚和器(aggregator)，透過將 `kube-aggregator` 佈署到 API Server 上，基於 `kubelet` 收集各個節點的指標數據再將數據儲存在 Metrics Server 的 Memory 中(代表不會保存歷史數據，重啟資料就會消失)，再以 API 的形式提供出來。

![https://ithelp.ithome.com.tw/upload/images/20220924/201495628ps2sa952O.png](https://ithelp.ithome.com.tw/upload/images/20220924/201495628ps2sa952O.png)

### 安裝 Metrics Server

接下來我們來把 Metrics Server 安裝到我們的 `Docker-Desktop` 上吧：

```jsx
#components.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    k8s-app: metrics-server
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
    rbac.authorization.k8s.io/aggregate-to-view: "true"
  name: system:aggregated-metrics-reader
rules:
  - apiGroups:
      - metrics.k8s.io
    resources:
      - pods
      - nodes
    verbs:
      - get
      - list
      - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    k8s-app: metrics-server
  name: system:metrics-server
rules:
  - apiGroups:
      - ""
    resources:
      - pods
      - nodes
      - nodes/stats
      - namespaces
      - configmaps
    verbs:
      - get
      - list
      - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server-auth-reader
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: extension-apiserver-authentication-reader
subjects:
  - kind: ServiceAccount
    name: metrics-server
    namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server:system:auth-delegator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
  - kind: ServiceAccount
    name: metrics-server
    namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-app: metrics-server
  name: system:metrics-server
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:metrics-server
subjects:
  - kind: ServiceAccount
    name: metrics-server
    namespace: kube-system
---
apiVersion: v1
kind: Service
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server
  namespace: kube-system
spec:
  ports:
    - name: https
      port: 443
      protocol: TCP
      targetPort: https
  selector:
    k8s-app: metrics-server
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: metrics-server
  strategy:
    rollingUpdate:
      maxUnavailable: 0
  template:
    metadata:
      labels:
        k8s-app: metrics-server
    spec:
      containers:
        - args:
            - --cert-dir=/tmp
            - --secure-port=4443
            - --kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname
            - --kubelet-use-node-status-port
            - --kubelet-insecure-tls # 加上這個  Do not verify the CA of serving certificates presented by Kubelets. For testing purposes only.
          image: k8s.gcr.io/metrics-server/metrics-server:v0.4.2
          imagePullPolicy: IfNotPresent
          livenessProbe:
            failureThreshold: 3
            httpGet:
              path: /livez
              port: https
              scheme: HTTPS
            periodSeconds: 10
          name: metrics-server
          ports:
            - containerPort: 4443
              name: https
              protocol: TCP
          readinessProbe:
            failureThreshold: 3
            httpGet:
              path: /readyz
              port: https
              scheme: HTTPS
            periodSeconds: 10
          securityContext:
            readOnlyRootFilesystem: true
            runAsNonRoot: true
            runAsUser: 1000
          volumeMounts:
            - mountPath: /tmp
              name: tmp-dir
      nodeSelector:
        kubernetes.io/os: linux
      priorityClassName: system-cluster-critical
      serviceAccountName: metrics-server
      volumes:
        - emptyDir: {}
          name: tmp-dir
---
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  labels:
    k8s-app: metrics-server
  name: v1beta1.metrics.k8s.io
spec:
  group: metrics.k8s.io
  groupPriorityMinimum: 100
  insecureSkipTLSVerify: true
  service:
    name: metrics-server
    namespace: kube-system
  version: v1beta1
  versionPriority: 100
```

可能有些聰明的朋友直接去照著官方安裝指示下載但卻失敗，其中一個原因可能是沒有開啟第 136 行的設定 `--kubelet-insecure-tls` ，此參數可以讓 Metrics Server 禁用證書驗證，畢竟我們在本地是不會有證書的。

建立 Metrics Server：

```jsx
kubectl apply -f ./components.yaml
-------
serviceaccount/metrics-server created
clusterrole.rbac.authorization.k8s.io/system:aggregated-metrics-reader created
clusterrole.rbac.authorization.k8s.io/system:metrics-server created
rolebinding.rbac.authorization.k8s.io/metrics-server-auth-reader created
clusterrolebinding.rbac.authorization.k8s.io/metrics-server:system:auth-delegator created
clusterrolebinding.rbac.authorization.k8s.io/system:metrics-server created
service/metrics-server created
deployment.apps/metrics-server created
apiservice.apiregistration.k8s.io/v1beta1.metrics.k8s.io created
```

成功的在 kube-system 中建立：

```jsx
kubectl get pods -n kube-system | grep metrics-server
--------
metrics-server-9f897d54b-l2rc4           1/1     Running   0               5m11s
```

### **顯示資源使用訊息**

`kubectl top` 可以查看節點和Pod的資源使用訊息, 包含`node`和`pod`兩個子命令, 可以顯示相關的資源佔用率。

```jsx
kubectl top node
------
NAME             CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%   
docker-desktop   215m         5%     5593Mi          71%
```

```jsx
kubectl top pods -n kube-system
-------
NAME                                     CPU(cores)   MEMORY(bytes)   
coredns-6d4b75cb6d-pp56z                 6m           0Mi             
coredns-6d4b75cb6d-qk4vm                 6m           0Mi             
etcd-docker-desktop                      28m          0Mi             
kube-apiserver-docker-desktop            59m          0Mi             
kube-controller-manager-docker-desktop   32m          0Mi             
kube-proxy-8j6xq                         1m           0Mi             
kube-scheduler-docker-desktop            6m           0Mi             
metrics-server-9f897d54b-l2rc4           8m           0Mi             
storage-provisioner                      4m           0Mi             
vpnkit-controller                        1m           0Mi
```

## 結論

有個這等資源指標收集工具，讓我們對資源監控的領域跨出了一大步，有了監控使我們可以更進一步的對整個集群依照指標做出更多自動化的設定，像是我們接下來要講到的 `AutoScaling` ，因為再使用他之前，建立好 Metrics Server 是他最基本的要求。


相關文章：
- [從異世界歸來的第二一天 - Kubernetes Resources(一) - Request/Limit](https://ithelp.ithome.com.tw/articles/10295419)
- [從異世界歸來的第二二天 - Kubernetes Resources(二) - Namespace](https://ithelp.ithome.com.tw/articles/10296200)

相關程式碼同時收錄在：

[https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day24](https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day24)

Reference

****[Kubernetes Metrics Server](https://kubernetes-sigs.github.io/metrics-server/)****

****[Kubernetes核心指标监控——Metrics Server](https://www.cnblogs.com/zhangmingcheng/p/15770672.html)****

****[資源指標 metrics-server](https://ithelp.ithome.com.tw/articles/10241138)****