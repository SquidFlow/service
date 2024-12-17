package reader

const (
	AppTypeHelm              string = "helm"
	AppTypeKustomize         string = "kustomize"
	AppTypeDirectory         string = "dir"
	AppTypeKsonnet           string = "ksonnet"
	AppTypeHelmMultiEnv      string = "helm-multiple-env"
	AppTypeKustomizeMultiEnv string = "kustomize-multiple-env"
)

type (
	AppSourceOption struct {
		Repo                 string
		TargetRevision       string
		Path                 string
		Submodules           bool
		ApplicationSpecifier *AppSourceSpecifier
	}
	// AppSourceSpecifier contains application-specific configuration
	AppSourceSpecifier struct {
		HelmManifestPath string
	}
)
