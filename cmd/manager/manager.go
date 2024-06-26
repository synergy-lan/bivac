package manager

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	bivacCmd "github.com/camptocamp/bivac/cmd"
	"github.com/camptocamp/bivac/internal/manager"
	"github.com/camptocamp/bivac/internal/utils"
	"github.com/camptocamp/bivac/pkg/volume"
)

var (
	server       manager.Server
	orchestrator string

	// Orchestrators is a copy of manager.Orchestrators which allows orchestrator
	// configuration from Cobra variables
	Orchestrators manager.Orchestrators

	dbPath           string
	resticForgetArgs string

	providersFile       string
	targetURL           string
	retryCount          int
	logServer           string
	agentImage          string
	whitelistVolumes    string
	blacklistVolumes    string
	whitelistAnnotation bool
	parallelCount       int
	refreshRate         string
	backupInterval      string
)
var envs = make(map[string]string)

var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Start Bivac backup manager",
	Run: func(cmd *cobra.Command, args []string) {
		volumesFilters := volume.Filters{
			Blacklist:           strings.Split(blacklistVolumes, ","),
			Whitelist:           strings.Split(whitelistVolumes, ","),
			WhitelistAnnotation: whitelistAnnotation,
		}

		o, err := manager.GetOrchestrator(orchestrator, Orchestrators)
		if err != nil {
			log.Errorf("failed to retrieve orchestrator: %s", err)
			return
		}

		if agentImage == "" {
			managerVersion := bivacCmd.BuildInfo.Version
			agentImage = fmt.Sprintf("ghcr.io/camptocamp/bivac:%s", utils.ComputeDockerAgentImage(managerVersion))
		}

		err = manager.Start(bivacCmd.BuildInfo, o, server, volumesFilters, providersFile, targetURL, logServer, agentImage, retryCount, parallelCount, refreshRate, backupInterval)
		if err != nil {
			log.Errorf("failed to start manager: %s", err)
			return
		}
	},
}

func init() {
	managerCmd.Flags().StringVarP(&server.Address, "server.address", "", "0.0.0.0:8182", "Address to bind on.")
	envs["BIVAC_SERVER_ADDRESS"] = "server.address"
	managerCmd.Flags().StringVarP(&server.PSK, "server.psk", "", "", "Pre-shared key.")
	envs["BIVAC_SERVER_PSK"] = "server.psk"

	managerCmd.Flags().StringVarP(&orchestrator, "orchestrator", "o", "", "Orchestrator on which Bivac should connect to.")
	envs["BIVAC_ORCHESTRATOR"] = "orchestrator"

	managerCmd.Flags().StringVarP(&Orchestrators.Docker.Endpoint, "docker.endpoint", "", "unix:///var/run/docker.sock", "Docker endpoint.")
	envs["BIVAC_DOCKER_ENDPOINT"] = "docker.endpoint"
	managerCmd.Flags().StringVarP(&Orchestrators.Docker.Network, "docker.network", "", "bridge", "Docker network.")
	envs["BIVAC_DOCKER_NETWORK"] = "docker.network"

	managerCmd.Flags().StringVarP(&Orchestrators.Cattle.URL, "cattle.url", "", "", "The Cattle URL.")
	envs["CATTLE_URL"] = "cattle.url"
	managerCmd.Flags().StringVarP(&Orchestrators.Cattle.AccessKey, "cattle.accesskey", "", "", "The Cattle access key.")
	envs["CATTLE_ACCESS_KEY"] = "cattle.accesskey"
	managerCmd.Flags().StringVarP(&Orchestrators.Cattle.SecretKey, "cattle.secretkey", "", "", "The Cattle secret key.")
	envs["CATTLE_SECRET_KEY"] = "cattle.secretkey"

	managerCmd.Flags().StringVarP(&Orchestrators.Kubernetes.Namespace, "kubernetes.namespace", "", "", "Namespace where you want to run Bivac.")
	envs["KUBERNETES_NAMESPACE"] = "kubernetes.namespace"
	managerCmd.Flags().BoolVarP(&Orchestrators.Kubernetes.AllNamespaces, "kubernetes.all-namespaces", "", false, "Backup volumes of all namespaces.")
	envs["KUBERNETES_ALL_NAMESPACES"] = "kubernetes.all-namespaces"
	managerCmd.Flags().StringVarP(&Orchestrators.Kubernetes.KubeConfig, "kubernetes.kubeconfig", "", "", "Path to your kuberconfig file.")
	envs["KUBERNETES_KUBECONFIG"] = "kubernetes.kubeconfig"
	managerCmd.Flags().StringVarP(&Orchestrators.Kubernetes.AgentServiceAccount, "kubernetes.agent-service-account", "", "", "Specify service account for agents.")
	envs["KUBERNETES_AGENT_SERVICE_ACCOUNT"] = "kubernetes.agent-service-account"
	managerCmd.Flags().StringVarP(&Orchestrators.Kubernetes.AgentLabelsInline, "kubernetes.agent-labels", "", "app=bivac", "Additional labels for agents.")
	envs["KUBERNETES_AGENT_LABELS"] = "kubernetes.agent-labels"
	managerCmd.Flags().StringVarP(&Orchestrators.Kubernetes.AgentAnnotationsInline, "kubernetes.agent-annotations", "", "", "Additional annotations for agents.")
	envs["KUBERNETES_AGENT_ANNOTATIONS"] = "kubernetes.agent-annotations"

	managerCmd.Flags().StringVarP(&resticForgetArgs, "restic.forget.args", "", "--group-by host --keep-daily 15 --prune", "Restic forget arguments.")
	envs["RESTIC_FORGET_ARGS"] = "restic.forget.args"

	managerCmd.Flags().StringVarP(&providersFile, "providers.config", "", "/providers-config.default.toml", "Configuration file for providers.")
	envs["BIVAC_PROVIDERS_CONFIG"] = "providers.config"

	managerCmd.Flags().StringVarP(&targetURL, "target.url", "r", "", "The target URL to push the backups to.")
	envs["BIVAC_TARGET_URL"] = "target.url"

	managerCmd.Flags().IntVarP(&retryCount, "retry.count", "", 0, "Retry to backup the volume if something goes wrong with Bivac.")
	envs["BIVAC_RETRY_COUNT"] = "retry.count"

	managerCmd.Flags().StringVarP(&logServer, "log.server", "", "", "Manager's API address that will receive logs from agents.")
	envs["BIVAC_LOG_SERVER"] = "log.server"

	managerCmd.Flags().StringVarP(&agentImage, "agent.image", "", "", "Agent's Docker image.")
	envs["BIVAC_AGENT_IMAGE"] = "agent.image"

	managerCmd.Flags().StringVarP(&whitelistVolumes, "whitelist", "", "", "Whitelist volumes.")
	envs["BIVAC_WHITELIST"] = "whitelist"
	envs["BIVAC_VOLUMES_WHITELIST"] = "whitelist"

	managerCmd.Flags().StringVarP(&blacklistVolumes, "blacklist", "", "", "Blacklist volumes.")
	envs["BIVAC_BLACKLIST"] = "blacklist"
	envs["BIVAC_VOLUMES_BLACKLIST"] = "blacklist"

	managerCmd.Flags().BoolVarP(&whitelistAnnotation, "whitelist.annotations", "", false, "Require pvc whitelist annotation")
	envs["BIVAC_WHITELIST_ANNOTATION"] = "whitelist.annotations"

	managerCmd.Flags().IntVarP(&parallelCount, "parallel.count", "", 2, "The count of agents to run in parallel")
	envs["BIVAC_PARALLEL_COUNT"] = "parallel.count"

	managerCmd.Flags().StringVarP(&refreshRate, "refresh.rate", "", "10m", "The volume list refresh rate.")
	envs["BIVAC_REFRESH_RATE"] = "refresh.rate"

	managerCmd.Flags().StringVarP(&backupInterval, "backup.interval", "", "23h", "Interval between two backups of a volume.")
	envs["BIVAC_BACKUP_INTERVAL"] = "backup.interval"

	bivacCmd.SetValuesFromEnv(envs, managerCmd.Flags())
	bivacCmd.RootCmd.AddCommand(managerCmd)
}
