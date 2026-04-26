---
name: platform
description: 언어·런타임·운영 대상 플랫폼 제약
---

# 플랫폼

## 빌드 / 런타임

- 언어: **Go 1.22**.
- 컨테이너: **distroless `nonroot`** (`gcr.io/distroless/static:nonroot`). 빌드 산출물은 `CGO_ENABLED=0` 정적 바이너리.
- 빌드 타깃: **`linux/amd64`** 단일. 다른 아키텍처 빌드는 추가 설계 필요 (Dockerfile 수정).

## 의존성

- controller-runtime: **v0.17.x** (현재 0.17.0).
- k8s.io/{api,apimachinery,client-go}: **v0.29.x** (현재 0.29.0).
- 의존 버전을 올릴 때는 controller-runtime 와 k8s.io/* 가 호환되는 짝을 함께 올린다 (`sigs.k8s.io/controller-runtime` README 의 호환 매트릭스 확인).

## 운영 대상 K8s

- **K8s ≥ 1.24** 기준. eviction policy/v1 GA, Lease 기반 leader-election 가정.
- 더 낮은 버전 지원이 필요하면 별도 설계 (eviction policy/v1beta1 fallback 등).
- `helm/Chart.yaml` `kubeVersion: ">=1.24.0-0"` 와 동기화 유지.

## 의존 추가 시 점검

- controller-runtime / k8s.io/* 외의 외부 SDK 추가는 보수적으로 — 컨트롤러 바이너리 크기와 RBAC 영향 검토.
- 새로 호출하는 K8s 리소스/verb 가 생기면 `helm/templates/clusterrole.yaml` 의 rules 와 `.claude/rules/subagents.md` 에 따라 보안 점검 후 함께 갱신.
