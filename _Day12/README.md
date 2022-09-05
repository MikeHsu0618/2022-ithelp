# 從異世界歸來的第十二天 - Kubernetes Deployment Strategies - 常見的部署策略 (一)

## 概述

如今在我們的實際工作環境中，產品的生命週期越來越短且迭代速度日益加快，身為 Server 守護者的我們可能會面臨到一兩週就要迭代一次，甚至是一天迭代數次的情況。我們需要知道大部分的技術創新價值始終來自於人們對它的了解以及使用程度，而有效的部署策略也因此是技術創新的一個重要因素，不論是推出時機、授權、行銷都是，一個適當的部署策略可以使我們一探市場接受程度、降低服務停運時間以及在機會成本中做出取捨。

在這邊可以跟大夥分享一個我做夢夢到的都市傳說，就是大家可能都疑惑過為合曾經的電商龍頭 PChome 首頁永遠都是十幾年前的 old school 風格(最近似乎又終於開始砍掉重練，疑我怎麼說又呢？)，難道 PChome 裡面沒有厲害的工程師？又或者是 **Legacy Code** 已經到無可救藥的地步？我倒覺得這兩種可能性較低，於是一位在 PChome 工作幾年的朋友解決了我多年的疑惑，原來曾經他們也有試著將首頁的 UI 改版，但由於改版的當天電商業績馬上掉好幾成只好緊急退版處理，從此他們的首頁變成沒有人敢打開的潘朵拉的盒子。

從上面的夢中內容我們可以觀察到一個好的部署策略應該是要具備盡可能的讓服務不中斷、可以戰略性的試試市場水溫並且需要擁有回退到歷史版本的能力，接下來就來介紹幾種常見的部署策略以及他們的優劣吧。

## 六種常見的部署策略

相信大家如果查詢過 `Kubernetes` 就容易看到以下的關鍵字眼，各種眼花撩亂的策略命名總會讓人第一眼無法參透其意，但了解其背後涵義又會覺得有幾分道理。需要特別強調這些部署方法並不是 `Kubernetes` 獨有，以任何方式實現該策略的概念精隨就可以說是執行了該部署策略，只是因為 `Kubernetes` 有容器管理、擴縮、調度多方面的強項，使得它能從容優雅的完成各種艱鉅的部署任務。下面就簡單的來介紹這幾種部署方式的區別以及優劣勢：

### 1. 重建部署 (Recreate)

重建部署是一個成本偏沈重的方式，簡單來說它會將舊版本完全下線後才開始上線新版本，這意味著你的服務將會有一段停機時間依賴於應用下線跟啟動的耗時。

優點：

- 便於設定。
- 線上只會同時運行一種版本。
- 部署過程中不會造成主機額外負擔。

缺點：

- 對使用者影響大，預期的停機時間取決於下線時間和啟動服務的耗時。

![https://storage.googleapis.com/cdn.thenewstack.io/media/2017/11/c42fa239-recreate.gif](https://storage.googleapis.com/cdn.thenewstack.io/media/2017/11/c42fa239-recreate.gif)

### 2. 滾動部署 (Ramped aka. Rolling-update)

滾動部署策略就是指容器會如同水漸漸地往傾斜的地方聚集一樣的更新版本，能緩慢平和的釋出新版本。

如同水流的流速快慢一樣，滾動部署也能通過調整下列引述來調整部署穩定性以及速率：

- 最大執行數：可以同時釋出的服務數目。
- 最大峰值：升級過程中最多可以比原先設定所多出的服務數量。
- 最大不可用數：最多可以有幾個服務處在無法服務的狀態。

優點：

- 相較於藍綠部署更加節省資源。
- 便於設定，服務不中斷。

缺點：

- 釋出與回滾耗時，想想當我們有 100 個服務，每次需要花五分鐘更新其中 10 個，當更新到第 80 個時發現錯誤需要緊急回滾的情況？
- 部署期間新舊兩版服務都會同時在線上運作，無法控制流量且噴錯時除錯困難高。

![https://storage.googleapis.com/cdn.thenewstack.io/media/2017/11/5bddc931-ramped.gif](https://storage.googleapis.com/cdn.thenewstack.io/media/2017/11/5bddc931-ramped.gif)

### 3. 藍綠部署 (Blue / Green)

相較於滾動更新，藍綠部署則是會先將新版本的服務完整的開啟，並且在新版本滿足上線條件的測試後，才將流量在負載均衡層從舊版本切換到新版本。

優點：

- 實時釋出、回滾。
- 避免新舊版本衝突，整個過程同時只會有一個版本存在。
- 服務不中斷。

缺點：

- 部署完成前需要雙倍的資源要求所增加的開銷及成本。有時新版本通過不了測試時，舊版本將持續運行到新版本通過為止。
- 當切換到新版本的瞬間，如果有未處理完成的業務將會是比較麻煩的問題。

![https://storage.googleapis.com/cdn.thenewstack.io/media/2017/11/73a2824d-blue-green.gif](https://storage.googleapis.com/cdn.thenewstack.io/media/2017/11/73a2824d-blue-green.gif)

### 4. 金絲雀部署 (Canary)

金絲雀釋出，與藍綠部署不同的是，它不是非黑即白的部署方式，所以又稱為灰度部署。

灰度釋出是指在黑與白之間，能夠平滑過度的一種部署方式。我們能夠緩慢的將新版本先推廣到一小部分的使用者，驗證沒有問題後才完成部署，以降低生產環境引入新功能帶來的風險。

例如將 90% 的請求導向舊版本，10% 的請求轉向新版本。這種部署大多用於缺少可靠測試或者對新版本穩定性缺乏信心的情況下。

<aside>
💡 金絲雀部署的命名來自於 17 世紀的礦井工人發現金絲雀對瓦斯這種氣體非常敏感，哪怕是只有及其微量的瓦斯，金絲雀也會停止歌唱率先比人類出現不良反應，所以工人每次下井時都會帶上一隻金絲雀作為危險狀況下的救命符。

</aside>

優點：

- 方便除錯以及監控。
- 只向一小部分使用者釋出。
- 快速回滾、快速迭代。

缺點：

- 完整釋出期漫長。
- 只能適用於相容迭代的方式，如果是大版本不相容就沒辦法使用這種方式了。

![https://storage.googleapis.com/cdn.thenewstack.io/media/2017/11/a6324354-canary.gif](https://storage.googleapis.com/cdn.thenewstack.io/media/2017/11/a6324354-canary.gif)

### 5. A / B 測試 (A / B Testing)

A / B 測試實際上是一種基於統計信息而非部署策略來製定業務決策的技術，與業務結合非常緊密。但是它們也是相關的，也可以使用金絲雀發布來實現。

除了基於權重在版本之間進行流量控制之外，A / B 測試還可以基於一些其他參數（比如Cookie、User Agent、地區等等）來精確定位給定的用戶群，該技術廣泛用於測試一些功能特性的效果，然後按照效果來進行確定。

A / B 測試是線上同時執行多個不同版本的服務，這些服務更多的是使用者側的體驗不同，比如頁面佈局、按鈕顏色，互動方式等，通常底層業務邏輯還是一樣的，也就是通常說的換湯不換藥。諸如 `Google Analysis` 等網站分析工具服務通常也可以搭配自家負載均衡器時現 A / B 測試。

優點：

- 多版本並行執行。
- 完全控制流量分佈。

缺點：

- 需要更全面的負載均衡(通常由雲端服務實現)。
- 難以定位辨別(通常由雲端服務實現分散式追蹤)。

![https://storage.googleapis.com/cdn.thenewstack.io/media/2017/11/5deeea9c-a-b.gif](https://storage.googleapis.com/cdn.thenewstack.io/media/2017/11/5deeea9c-a-b.gif)

### 6. 影子部署 (Shadow)

影子部署是指在原有版本旁完整運行新版本，並且將流入原有版本的請求同時分發到新版本，得以實現在更新之前就模擬正式產品環境的運作情況，直到滿足上線條件後才將進入點轉往新版本並關閉舊版本。

非常理想的流程但背後所需要時現的技術門檻與成本相當的高，尤其需要特別注意特殊情況下的無法掌控的狀況。例如一個下單的請求同時轉向新舊版本的服務，最終可能導致下單兩次的結果。

優點：

- 可以直接對正式環境流量進行效能測試而不影響使用者。
- 直到應用穩定且達到上線條件時才釋出。

缺點：

- 與藍綠部署一樣需要雙倍的資源請求。
- 配置複雜，容易出現預期外的情況。

![https://storage.googleapis.com/cdn.thenewstack.io/media/2017/11/fdd947f8-shadow.gif](https://storage.googleapis.com/cdn.thenewstack.io/media/2017/11/fdd947f8-shadow.gif)

## 結論

很高興這些曾經看似熟悉卻說不出個所以然的花裡胡哨的策略，可以藉由這次的機會讓我一次將他們做個完整的介紹，但是只知道原理是遠遠不夠的～我們可是還沒發揮到 `Kubernetes` 在容器調度以及部署方面的強項呢，所以接下來的幾天我們即將會回到歡樂的 `Kubernetes` 實戰環節，請大家多多擔待啦。

Reference

**[超詳細GA網站分析入門大全，看這篇就對了！](https://www.webguide.nat.gov.tw/News_Content.aspx?n=531&s=2935)**

[https://www.gushiciku.cn/pl/gUOs/zh-tw](https://www.gushiciku.cn/pl/gUOs/zh-tw)

**[5 Kubernetes Deployment Strategies: Roll Out Like the Pros](https://spot.io/resources/kubernetes-autoscaling/5-kubernetes-deployment-strategies-roll-out-like-the-pros/?utm_campaign=spot.io-ps-kubernatics&utm_term=&utm_source=adwords&utm_medium=ppc&hsa_ver=3&hsa_kw=&hsa_cam=14123334381&hsa_tgt=dsa-1510472120825&hsa_acc=8916801654&hsa_mt=&hsa_net=adwords&hsa_ad=567046495853&hsa_src=g&hsa_grp=129315742943&gclid=CjwKCAjw9suYBhBIEiwA7iMhNDpvsUblnJVcCphnsMwrYLjUJQTSe9_mK1naBRa1aRtibTx0yZTrEhoCjaUQAvD_BwE)**

****[Six Strategies for Application Deployment](https://thenewstack.io/deployment-strategies/)****