從異世界歸來的第十七天 - Kubernetes Volume (二) - EmptyDir
---

## 概述

在我們前篇文章中簡單的介紹了 `emptyDir` 並提到他與 `Pod` 的生命週期共生共滅，所以他通常被用於數據緩存或者臨時存儲的場景，接下來就來實際操作練習一下吧。

### EmptyDir Volume

`emptyDir` 可以簡單理解為在 `Pod` 運作時的一個臨時目錄，就像我們在 `Docker` 上使用的 `Volume` ，而在 `Pod` 被移除時會被一併刪除，除了一些特殊場景通常我們都會將他用於在一個 `Pod` 內的多個容器間的文件的共享，或做為容器數據的臨時緩存存儲目錄等等。

emptyDir存儲卷則定義於. **spec.volumes.emptyDir** 嵌套字段中，可用字段主要包含兩個，具體如下：

- **medium**：此目錄所在存儲介質的類型，可取值為**default**或**Memory**，默認為default，表示使用節點的默認存儲介質：**Memory**表示基於RAM的臨時文件系統tmpfs，空間受於內存，但性能非常好，通常用於為容器中的應用提供緩存空間。
- **sizeLimit：**當前存儲卷的空間限額，默認值為nil，表示不限制；不過在medium 字段為**Memory**時，建議定義此限額。

### 1. 創建一個有多個容器的 Pod

```jsx
apiVersion: v1
kind: Pod
metadata:
  name: emptydir-pod
spec:
  volumes:
    - name: html
      emptyDir: {}
  containers:
    - name: nginx
      image: nginx:latest
      volumeMounts:
        - name: html
          mountPath: /usr/share/nginx/html
    - name: alpine
      image: alpine
      volumeMounts:
        - name: html
          mountPath: /html
      command: [ "/bin/sh", "-c" ]
      args: # 每十秒定時向 /html/index.html 寫入資料
        - while true; do
          echo $(hostname) $(date) >> /html/index.html;
          sleep 10;
          done
```

我們可以很簡單的看到再 Pod 的設定檔中，我們啟用了 `nginx` 以及 `alpine` 並且掛載同一個的 `emptyDir` 的共享目錄

### 2. 創建 Pod

```jsx
kubectl apply -f emptydir.yaml
--------

pod/emptydir-pod created
```

### 3. 查看 Pod 狀態

```jsx
kubectl describe pod/emptydir-pod
--------
Containers:
  nginx:
    Container ID:   docker://64883c01c4e987beaa4cfbda1bba5cbe571b934dcc47b978e4adca4569a21170
    Image:          nginx:latest
    Image ID:       docker-pullable://nginx@sha256:1761fb5661e4d77e107427d8012ad3a5955007d997e0f4a3d41acc9ff20467c7
    Port:           <none>
    Host Port:      <none>
    State:          Running
      Started:      Tue, 26 Jul 2022 17:05:41 +0800
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /usr/share/nginx/html from html (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-bmdwh (ro)
  alpine:
    Container ID:  docker://7881eb57f048f56e9d9ed4eedab818aaf876138cc488bef79746008c2a1047e9
    Image:         alpine
    Image ID:      docker-pullable://alpine@sha256:7580ece7963bfa863801466c0a488f11c86f85d9988051a9f9c68cb27f6b7872
    Port:          <none>
    Host Port:     <none>
    Command:
      /bin/sh
      -c
    Args:
      while true; do echo $(hostname) $(date) >> /html/index.html; sleep 10; done
    State:          Running
      Started:      Tue, 26 Jul 2022 17:05:47 +0800
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /html from html (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-bmdwh (ro)
	Conditions:
	  Type              Status
	  Initialized       True 
	  Ready             True 
	  ContainersReady   True 
	  PodScheduled      True 
	Volumes:
	  html:
	    Type:       EmptyDir (a temporary directory that shares a pod's lifetime)
	    Medium:     
	    SizeLimit:  <unset>
```

可以從返回結果中看到各個容器與 `Volume` 的掛載狀態。

### 4. 訪問 Pod 中的 Nginx

來確認一下 `alpine` 容器每隔 10 秒像 html/index.html 寫入訊息，而 `Nginx` 容器掛載的 `emptyDir` 是否同時也可以取得更新。

將 port 導出到本地的 [localhost](http://localhost) :

```jsx
kubectl port-forward pod/emptydir-pod 8080:80
-------
Forwarding from 127.0.0.1:8080 -> 80pod 8080:80
Forwarding from [::1]:8080 -> 80
```

使用 curl 查看返回值：

```jsx
curl http://localhost:8080
--------
emptydir-pod Tue Jul 26 09:05:47 UTC 2022
emptydir-pod Tue Jul 26 09:05:57 UTC 2022
emptydir-pod Tue Jul 26 09:06:07 UTC 2022
emptydir-pod Tue Jul 26 09:06:17 UTC 2022
emptydir-pod Tue Jul 26 09:06:27 UTC 2022
```

順利取得由 `alpine` 容器產生的內容～

### 5. 進入容器查看

通過 -c 可以指定容器名稱進入指定容器

```jsx
kubectl exec -it pods/emptydir-pod -c nginx -- sh

head -3 /usr/share/nginx/html/index.html
--------
emptydir-pod Tue Jul 26 09:05:47 UTC 2022
emptydir-pod Tue Jul 26 09:05:57 UTC 2022
emptydir-pod Tue Jul 26 09:06:07 UTC 2022
```

```jsx
kubectl exec -it pods/emptydir-pod -c alpine -- sh

head -3 /html/index.html
--------
emptydir-pod Tue Jul 26 09:05:47 UTC 2022
emptydir-pod Tue Jul 26 09:05:57 UTC 2022
emptydir-pod Tue Jul 26 09:06:07 UTC 2022

ps aux
--------
PID   USER     TIME  COMMAND
    1 root      0:00 /bin/sh -c while true; do echo $(hostname) $(date) >> /html/index.html; sleep 10; done
  371 root      0:00 sh
  395 root      0:00 sleep 10
  396 root      0:00 ps aux
```

### 6. 設定 Memory 作為高性能緩存

```jsx
apiVersion: v1
kind: Pod
metadata:
  name: emptydir-memory-pod
spec:
  volumes:
    - name: html
      emptyDir:
        medium: Memory                #指定使用記憶體儲存
        sizeLimit: 256Mi              #限制內存大小
  containers:
    - name: nginx
      image: nginx:latest
      volumeMounts:
        - name: html
          mountPath: /usr/share/nginx/html
    - name: alpine
      image: alpine
      volumeMounts:
        - name: html
          mountPath: /html
      command: [ "/bin/sh", "-c" ]
      args:
        - while true; do
          echo $(hostname) $(date) >> /html/index.html;
          sleep 10;
          done
```

## 結論

上面簡單的例子充分的展現出 `emptyDir` 特有的定位以及簡單易懂的用法，但這些完全只是 `Kubernetes Volumes` 中多種類的其中之一而已，隨後還會有幾篇文章來介紹其他常用的 `Volume` 。

相關文章：

- [從異世界歸來的第十六天 - Kubernetes Volume (一) - Volume 是什麼](https://ithelp.ithome.com.tw/articles/10291557)

相關程式碼同時收錄在：

[https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day17](https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day17)

Reference

****[Kubernetes中的emptyDir存儲捲和節點存儲卷](https://cloud.tencent.com/developer/article/1660415)**** 