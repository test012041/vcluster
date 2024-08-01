package e2evclusterinstall

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/loft-sh/vcluster/pkg/platform/random"
	"github.com/loft-sh/vcluster/test/framework"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

const (
	pollingInterval     = time.Second * 2
	pollingDurationLong = time.Minute * 2
	vclusterRepo        = "https://charts.loft.sh"
	filePath            = "../commonValues.yaml"
	outputFile          = "output.yaml"
)

var _ = ginkgo.Describe("Deploy and Delete vCluster", func() {
	ginkgo.It("should deploy a vCluster using kubectl and delete it using kubectl", func() {
		vClusterName := "t-cluster-" + random.String(6)
		vClusterNamespace := "t-ns-" + random.String(6)
		file, err := os.Create(outputFile)
		framework.ExpectNoError(err)
		defer file.Close()

		createNamespaceCmd := exec.Command("kubectl", "create", "namespace", vClusterNamespace)
		err = createNamespaceCmd.Run()
		framework.ExpectNoError(err)

		helmCmd := exec.Command("helm", "template", vClusterName, "vcluster", "--repo", vclusterRepo, "-n", vClusterNamespace, "-f", filePath)
		helmCmd.Stdout = file
		err = helmCmd.Run()
		framework.ExpectNoError(err)

		kubectlCmd := exec.Command("kubectl", "apply", "-f", outputFile)
		err = kubectlCmd.Run()
		framework.ExpectNoError(err)

		gomega.Eventually(func() bool {

			checkCmd := exec.Command("vcluster", "list")
			output, err := checkCmd.CombinedOutput()
			framework.ExpectNoError(err)
			return err == nil && strings.Contains(string(output), vClusterName) && strings.Contains(string(output), "Running")
		}).WithPolling(pollingInterval).WithTimeout(pollingDurationLong).Should(gomega.BeTrue())

		deleteCmd := exec.Command("kubectl", "delete", "-f", outputFile)
		err = deleteCmd.Run()
		framework.ExpectNoError(err)

		gomega.Eventually(func() bool {

			checkCmd := exec.Command("vcluster", "list")
			output, err := checkCmd.CombinedOutput()
			framework.ExpectNoError(err)
			return err == nil && !strings.Contains(string(output), vClusterName)
		}).WithPolling(pollingInterval).WithTimeout(pollingDurationLong).Should(gomega.BeTrue())

	})
})
