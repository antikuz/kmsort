package main

var ignoreSpecValues = map[string][]string{
	"Service": {"clusterIP","clusterIPs"},
}

var orderMap = map[string][]string{
	"Service":        {"selector", "ports"},
	"Deployment":     {"replicas", "selector", "nodeSelector", "template", "strategy"},
	"StatefulSet":    {"replicas", "selector", "template"},
	"metadata":       {"name", "namespace", "labels", "annotations"},
	"containers":     {"name", "image", "imagePullPolicy", "ports"},
	"initContainers": {"name", "image", "imagePullPolicy", "ports"},
	"volumeMounts":   {"name"},
	"volumes":        {"name"},
	"configMap":      {"name"},
	"Template":       {"nodeSelector", "image", "initContainers", "containers"},
	"strategy":       {"type"},
	"updateStrategy": {"type"},
	"ports":          {"name", "port", "containerPort", "protocol"},
	"spec":           {"nodeSelector", "image", "initContainers", "containers"},
}

var defaultSpecValues = map[string]interface{}{
	"internalTrafficPolicy":         "Cluster",
	"type":                          "ClusterIP",
	"ipFamilies":                    []interface{}{"IPv4"},
	"ipFamilyPolicy":                "SingleStack",
	"sessionAffinity":               "None",
	"revisionHistoryLimit":          10,
	"progressDeadlineSeconds":       600,
	"dnsPolicy":                     "ClusterFirst",
	"restartPolicy":                 "Always",
	"terminationGracePeriodSeconds": 30,
	"terminationMessagePolicy":      "File",
	"terminationMessagePath":        "/dev/termination-log",
	"failureThreshold":              3,
	"successThreshold":              1,
	"timeoutSeconds":                1,
	"periodSeconds":                 10,
	"initialDelaySeconds":           0,
	"defaultMode":                   420,
	"readOnlyRootFilesystem":        false,
	"runAsNonRoot":                  false,
}

var fieldsPopulatedByTheSystem = []string{
	"generateName",
	"selfLink",
	"uid",
	"resourceVersion",
	"generation",
	"creationTimestamp",
	"deletionTimestamp",
	"deletionGracePeriodSeconds",
	"status",
}
