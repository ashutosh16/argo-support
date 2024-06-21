curl -v 'http://localhost:4000/api/v1/applications/argo-rollouts/resource/actions?appNamespace=argocd&namespace=testns&resourceName=istio-host-split&version=v1alpha1&kind=Rollout&group=argoproj.io' \
  -H 'Accept: */*' \
  -H 'Accept-Language: en-US,en;q=0.9' \
  -H 'Cache-Control: no-cache' \
  -H 'Connection: keep-alive' \
  -H 'Content-Type: application/json' \
  -H 'Cookie: argocd.token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhcmdvY2QiLCJzdWIiOiJhZG1pbjpsb2dpbiIsImV4cCI6MTcxMDU0MDg5MywibmJmIjoxNzEwNDU0NDkzLCJpYXQiOjE3MTA0NTQ0OTMsImp0aSI6ImQ3YzBkYmNlLTM5NmQtNDIyNi1hNjljLWFkNDNlMTQ4MTk0ZiJ9.EjmfXreJhnxDdN_rwfupil7RuEtDLG_xwXzn_-5z9D0' \
  -H 'Origin: http://localhost:4000' \
  -H 'Pragma: no-cache' \
  -H 'Referer: http://localhost:4000/applications/argocd/test2?view=tree&resource=&extension=AI' \
  -H 'Sec-Fetch-Dest: empty' \
  -H 'Sec-Fetch-Mode: cors' \
  -H 'Sec-Fetch-Site: same-origin' \
  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36' \
  -H 'sec-ch-ua: "Google Chrome";v="123", "Not:A-Brand";v="8", "Chromium";v="123"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "macOS"' \
  --data-raw '"create-genai"'


  curl 'http://localhost:4000/api/v1/applications/argo-rollouts/resource?name=istio-host-split&appNamespace=argocd&namespace=testns1&resourceName=istio-host-split&version=v1alpha1&kind=Rollout&group=argoproj.io&patchType=application%2Fmerge-patch%2Bjson' \
    -H 'Accept: */*' \
    -H 'Accept-Language: en-US,en;q=0.9' \
    -H 'Connection: keep-alive' \
    -H 'Content-Type: application/json' \
    -H 'Cookie: Goland-3cd825d8=e8ac4b08-5f83-47fb-ba23-29044f6ab219; argocd.token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhcmdvY2QiLCJzdWIiOiJhZG1pbjpsb2dpbiIsImV4cCI6MTcxNTcxNDEzNCwibmJmIjoxNzE1NjI3NzM0LCJpYXQiOjE3MTU2Mjc3MzQsImp0aSI6IjdmMjQ1ZTIxLTMxM2MtNGI1OS1iMWE0LThiY2JiZDBkZmY0NSJ9.mK5lzMPU8pFIuuSJeHnhoO7wQaxgbs-C_QmSVJJROcQ' \
    -H 'Origin: http://localhost:4000' \
    -H 'Referer: http://localhost:4000/applications/argocd/argo-rollouts?resource=&node=argoproj.io%2FRollout%2Ftestns1%2Fistio-host-split%2F0' \
    -H 'Sec-Fetch-Dest: empty' \
    -H 'Sec-Fetch-Mode: cors' \
    -H 'Sec-Fetch-Site: same-origin' \
    -H 'User-Agent: Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36' \
    -H 'sec-ch-ua: "Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"' \
    -H 'sec-ch-ua-mobile: ?argocd-cm.yaml' \
    -H 'sec-ch-ua-platform: "Android"' \
    --data-raw '"{\"status\":{\"observedGeneration\":null}}"'