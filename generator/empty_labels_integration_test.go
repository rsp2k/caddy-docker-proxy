package generator

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

func TestContainers_EmptyCaddyLabel(t *testing.T) {
	dockerClient := createBasicDockerClientMock()
	dockerClient.ContainersData = []types.Container{
		{
			ID: "container-with-empty-label",
			Names: []string{
				"problematic-container",
			},
			NetworkSettings: &types.SummaryNetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"caddy-network": {
						IPAddress: "172.17.0.2",
						NetworkID: caddyNetworkID,
					},
				},
			},
			Labels: map[string]string{
				"caddy":               "", // Empty label that should be ignored
				"caddy.reverse_proxy": "172.17.0.2:5000",
			},
		},
	}

	// Should generate Caddyfile with remaining valid labels (but no site block due to missing main caddy label)
	const expectedCaddyfile = "{\n" +
		"	reverse_proxy 172.17.0.2:5000\n" +
		"}\n"

	// Should contain error log about the empty label
	const expectedLogs = commonLogs +
		"ERROR	🚨 IGNORING EMPTY CADDY LABEL - This would break all proxied sites!	" +
		"{\"container_id\": \"container-with-empty-label\", \"problematic_label\": \"caddy\", \"label_value\": \"''\", \"action\": \"skipping_container_label\"}\n"

	testGeneration(t, dockerClient, nil, expectedCaddyfile, expectedLogs)
}

func TestContainers_WhitespaceOnlyCaddyLabel(t *testing.T) {
	dockerClient := createBasicDockerClientMock()
	dockerClient.ContainersData = []types.Container{
		{
			ID: "container-with-whitespace-label",
			Names: []string{
				"whitespace-container",
			},
			NetworkSettings: &types.SummaryNetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"caddy-network": {
						IPAddress: "172.17.0.3",
						NetworkID: caddyNetworkID,
					},
				},
			},
			Labels: map[string]string{
				"caddy":               "  \n\t  ", // Whitespace-only label
				"caddy.reverse_proxy": "172.17.0.3:5000",
			},
		},
	}

	// Should generate Caddyfile with remaining valid labels (but no site block due to missing main caddy label)
	const expectedCaddyfile = "{\n" +
		"	reverse_proxy 172.17.0.3:5000\n" +
		"}\n"

	// Should contain error log about the whitespace-only label
	const expectedLogs = commonLogs +
		"ERROR	🚨 IGNORING EMPTY CADDY LABEL - This would break all proxied sites!	" +
		"{\"container_id\": \"container-with-whitespace-label\", \"problematic_label\": \"caddy\", \"label_value\": \"'  \\n\\t  '\", \"action\": \"skipping_container_label\"}\n"

	testGeneration(t, dockerClient, nil, expectedCaddyfile, expectedLogs)
}

func TestContainers_EmptyNestedLabels(t *testing.T) {
	dockerClient := createBasicDockerClientMock()
	dockerClient.ContainersData = []types.Container{
		{
			ID: "container-with-empty-nested-labels",
			Names: []string{
				"nested-empty-container",
			},
			NetworkSettings: &types.SummaryNetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"caddy-network": {
						IPAddress: "172.17.0.4",
						NetworkID: caddyNetworkID,
					},
				},
			},
			Labels: map[string]string{
				"caddy":               "nested.testdomain.com",
				"caddy.reverse_proxy": "172.17.0.4:5000",
				"caddy.tls":          "", // Valid empty directive
				"caddy.header":       "", // Valid empty directive
			},
		},
	}

	// Should generate Caddyfile with all valid directives (including valid empty ones)
	const expectedCaddyfile = "nested.testdomain.com {\n" +
		"	header\n" +
		"	reverse_proxy 172.17.0.4:5000\n" +
		"	tls\n" +
		"}\n"

	// Should NOT contain error logs since these are valid empty directives
	const expectedLogs = commonLogs

	testGeneration(t, dockerClient, nil, expectedCaddyfile, expectedLogs)
}

func TestContainers_MixedEmptyAndValidLabels(t *testing.T) {
	dockerClient := createBasicDockerClientMock()
	dockerClient.ContainersData = []types.Container{
		{
			ID: "container-with-mixed-labels",
			Names: []string{
				"mixed-container",
			},
			NetworkSettings: &types.SummaryNetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"caddy-network": {
						IPAddress: "172.17.0.5",
						NetworkID: caddyNetworkID,
					},
				},
			},
			Labels: map[string]string{
				"caddy":               "mixed.testdomain.com",
				"caddy.reverse_proxy": "172.17.0.5:5000",
				"caddy.tls":          "internal", // Valid label
				"caddy.respond":      "",         // Valid empty directive
				"caddy.header":       "X-Test-Header value", // Valid label
			},
		},
	}

	// Should generate Caddyfile with all valid directives (including valid empty respond)
	const expectedCaddyfile = "mixed.testdomain.com {\n" +
		"	header X-Test-Header value\n" +
		"	respond\n" +
		"	reverse_proxy 172.17.0.5:5000\n" +
		"	tls internal\n" +
		"}\n"

	// Should NOT contain error logs since respond is a valid empty directive
	const expectedLogs = commonLogs

	testGeneration(t, dockerClient, nil, expectedCaddyfile, expectedLogs)
}

func TestContainers_MultipleContainersWithEmptyLabels(t *testing.T) {
	dockerClient := createBasicDockerClientMock()
	dockerClient.ContainersData = []types.Container{
		{
			ID: "good-container",
			Names: []string{
				"good-container",
			},
			NetworkSettings: &types.SummaryNetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"caddy-network": {
						IPAddress: "172.17.0.6",
						NetworkID: caddyNetworkID,
					},
				},
			},
			Labels: map[string]string{
				"caddy":               "good.testdomain.com",
				"caddy.reverse_proxy": "172.17.0.6:5000",
			},
		},
		{
			ID: "bad-container",
			Names: []string{
				"bad-container",
			},
			NetworkSettings: &types.SummaryNetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"caddy-network": {
						IPAddress: "172.17.0.7",
						NetworkID: caddyNetworkID,
					},
				},
			},
			Labels: map[string]string{
				"caddy":               "", // This should not break the good container
				"caddy.reverse_proxy": "172.17.0.7:5000",
			},
		},
	}

	// Should generate Caddyfile for good container and global directives from bad container
	const expectedCaddyfile = "{\n" +
		"	reverse_proxy 172.17.0.7:5000\n" +
		"}\n" +
		"good.testdomain.com {\n" +
		"	reverse_proxy 172.17.0.6:5000\n" +
		"}\n"

	// Should contain error log about the bad container
	const expectedLogs = commonLogs +
		"ERROR	🚨 IGNORING EMPTY CADDY LABEL - This would break all proxied sites!	" +
		"{\"container_id\": \"bad-container\", \"problematic_label\": \"caddy\", \"label_value\": \"''\", \"action\": \"skipping_container_label\"}\n"

	testGeneration(t, dockerClient, nil, expectedCaddyfile, expectedLogs)
}