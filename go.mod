module github.com/dmage/openshift-trt

go 1.16

require (
	github.com/openshift/ci-tools v0.0.0-20210612121943-3cc79c4d3ea1
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/openshift/builder => github.com/openshift/builder v4.0.0+incompatible
	github.com/openshift/library-go => github.com/openshift/library-go v0.0.0-20200421122923-c1de486c7d47
	k8s.io/api => k8s.io/api v0.20.7
	k8s.io/client-go => k8s.io/client-go v0.20.7
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.7.0
)
