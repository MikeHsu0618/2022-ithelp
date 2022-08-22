# Day 20 Kubernetes AutoScaling (一)

## 概述

在前面的幾天裡，我們認識了很多關於資源配置以及監控的觀念，但如果我們掌握了這些資源指標卻只能手動調整就感覺失去了靈魂一樣，於是就有了自動化資源配置的 `AutoScaling` 出現， `AutoScaling` 簡單來說就是圍繞在你設定的資源配置並監控資源使用率去對系統做出水平、垂直、多維拓展或縮減來應對系統負載的波動，但需要再次強調的是，這一切都是建立在已經設定好資源配置以及 Metrics Server 為前提才能實現～。

## Autoscalers 的種類

### Cluster Autoscaler（CA）

`Cluster Autoscaler` 最主要的工作就是調節 node-pool 的數量，是屬於 cluster level 的 autoscaler，簡單來說他能幫我們在負載高時開新的 node，負載低時關閉 node。

- Scale-up：有 `Pod` 的狀態為 `unschedulable` 的十秒左右時將會迅速的判斷是否需要縱向擴展，需要注意的是縱向擴展的動作可以在十秒內左右完成，而開啟機器的時間可能需要數分鐘到十來分鐘才能處於可用狀態。
- Scale-down：每隔一段時間(預設為十秒)檢查 CPU 以及內存請求總和是否低於 50 %，並且沒有 `Pod` 或 `Node` 調度條件限制。
- 部分設定設不好會讓 CA 沒辦法自動擴展
    - CA 要關 node 然後 evict pod 時違反 Pod affinity/anti-affinity 和 PodDisruptionBudget
    - 在 node 加上 annotation 可防止被 scale down：`"cluster-autoscaler.kubernetes.io/scale-down-disabled": "true"`

就像是上面有提到的 `Cluster Auotscaler` 是屬於集群等級的調節者，所以我們在本地只有一個 node 的 `docker-desktop` 是沒辦法實際體驗到他的厲害之處，如果我們使用的是平台等級的 `Kubernetes` ，像是 `Gcp GKE` `Aws EKS` 等等，此類雲端平台整合更全面的資源並做出更高層級的自動調度。

### Horizontal Pod Autoscaler（HPA）

![horizontal-pod-autoscaler.svg](horizontal-pod-autoscaler.svg)

調節 `Pod` 數量的 autoscaler，屬於 pod level ，負責在負載波動時自動擴展或縮減 `Pod` 到設定的最大數量以及最小數量。

- **Scale-up：** 檢查 metrics server，發現過了設定使用率就增加 deployment 的 replicas。
- **Scale-down：** 檢查 metrics，發現過了設定使用率就減少 deployment 的 replicas。
- 如果 deployment 原本就有設定 replica 數目將會被 HPA 的 replica 設定覆蓋，等於是 HPA 將會無視 deployment 設定對 Pod 數量擴展縮減。
- 進行擴展或縮減後都會等待三到五分鐘到系統穩定後，再開始檢查 metrics server。
- 使用率計算：如果 `currentMetricValue`是 200m ，而 `desiredMetricValue`
  是 100m，則代表目前要增加 200/100 = 2 倍的 replica 的數量。

    ```jsx
    desiredReplicas = ceil[currentReplicas * ( currentMetricValue / desiredMetricValue )]
    ```

- 可以設定 custom/external metrics 來觸發 autoscaling。
- v2beta2 以上的 HPA 才有內存可以檢查，v1 只能檢查 CPU utilization。

<aside>
💡 目前 HPA 的 Api 更新的非常快，網路上的範例可以找到 v1、v2、v2beta2…等版本的範例教學，但只有從 v2beta2 開始才支援 metric 內存，所以還是建議多翻閱最新的 API 文件。

</aside>

### Vertical Pod Autoscaler（VPA）

自動推薦並設定最適合 pod resource requests/limits 的設定，簡單來說就是不再需要透過人工監控並且手動設定 CPU 和內存配置，可以替沒有系統調優相關經驗的 DevOps 小白鬆了一口氣。

- **Scale-up：** 檢查 metrics，發現過了設定使用率就減少 deployment 的 pod 的 resources.requests，再透過重啟 Pod 實際更新。
- **Scale-down：** 檢查 metrics，發現過了設定使用率就減少 deployment 的 Pod 的 resources.requests，再透過重啟 Pod 實際更新。
- 沒錯 HPA 的 `Updater` 再檢查到該 Pod 需要更新時，必須要將原本的 Pod 刪除並重新建立一個更新過 reqeusts/limits 的 Pod。
- VPA 的改動會參考 Pod 的 `history data`。
- 目前 VPA 還不能跟 HPA 一起混用，除非 HPA 使用得是 custom metric 當作 trigger，亦或者是！使用了下面介紹的 MPA（多維 Pod 自動擴縮）可以達到同時進行 VPA 和 HPA。

<aside>
💡 VPA 跟 Metrics Server 一樣都是 `custom resources` ，代表他們不一定會預設安裝在 `Kubernetes` 當中。但很多核心功能都仰賴這些 `custom resources` 來實現，這一切都讓 `Kubernetes` 更加模組化。

</aside>

### **Multidim Pod Autoscaler**（MPA）

MPA (多維 Pod 自動擴縮) 可以讓我們同時選擇多種方法來擴縮集群，目前只有 `GCP GKE` 有提供這種多維度的擴縮操作，在偉栽 Google 爸爸的之外也體會到了，因為 `Kubernetes` 開源社群中的 `autoscaler` 專案還沒實現多維擴縮 ，所以各大雲端平台都不支持並且都在等開源社群生出新功能再包進自己的雲端平台裡，非常的邪惡～

- 再次強調這功能目前當下只有 `GCP GKE` 限定，無法在本地或其他平台中實現。
- 目前此功能在 `GCP GKE` 屬於 beta 版本，有興趣可以在非正式環境中玩玩看。
- 多維度擴縮目前只能 `『根據 CPU 進行 HPA，以及根據內存進行 VPA』` ，所以要注意一下這邊不同於 VPA 的設定在於 `在 deployment 的 resources 中的 cpu requests/limits 是必須要事先設定的欄位，因為目前的 VPA 只能依據內存進行 VPA。

有興趣的可以參考參考[官方文件](https://cloud.google.com/kubernetes-engine/docs/how-to/multidimensional-pod-autoscaling)來玩玩看。

## 結論

大略介紹完了 `AutoScaling` 的種類，可以發現不管到哪裡資源的監控以及配置都是必修課題，畢竟沒有一個老闆願意多花一分冤枉錢，同時我們也可以觀察到 `Kubernetes` 在這一部分很多地方都需要仰賴 `custom resources` 來實現，並且這些功能都不能說到完善的階段，迭代是非常非常的快，只能送他老話一句：『別再更新了，老子學不動啦！』

Reference

**[Configuring multidimensional Pod autoscaling](https://cloud.google.com/kubernetes-engine/docs/how-to/multidimensional-pod-autoscaling)**

**[Scale container resource requests and limits](https://cloud.google.com/kubernetes-engine/docs/how-to/vertical-pod-autoscaling)**

[kubernetes](https://github.com/kubernetes)/**[autoscaler](https://github.com/kubernetes/autoscaler)**

**[Kubernetes Autoscaling: 3 Methods and How to Make Them Great](https://spot.io/resources/kubernetes-autoscaling-3-methods-and-how-to-make-them-great/?utm_campaign=spot.io-ps-kubernatics&utm_term=&utm_source=adwords&utm_medium=ppc&hsa_ver=3&hsa_kw=&hsa_cam=14123334381&hsa_tgt=dsa-1510472121065&hsa_acc=8916801654&hsa_mt=&hsa_net=adwords&hsa_ad=567046495853&hsa_src=g&hsa_grp=129315742943&gclid=Cj0KCQjwr4eYBhDrARIsANPywCgPp5OARQRTY2SAzYqqKU8hIfsZKY3SIwCabTNBG31vFpKzbTijzycaAk92EALw_wcB#a3)**

****[Kubernetes Horizontal Scaling/Vertical Scaling 概念](https://sean22492249.medium.com/kubernetes-horizontal-scaling-vertical-scaling-%E6%A6%82%E5%BF%B5-e8e70ce6f034)****

**[Kuberenetes Autoscaling 相關知識小整理](https://weihanglo.tw/posts/2020/k8s-autoscaling/)**

****[Kubernetes 那些事 — Auto Scaling](https://medium.com/andy-blog/kubernetes-%E9%82%A3%E4%BA%9B%E4%BA%8B-auto-scaling-7b887f61fdec)****