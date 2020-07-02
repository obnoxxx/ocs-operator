package defaults

import (
	rook "github.com/rook/rook/pkg/apis/rook.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// appLabelSelectorKey is common value for 'Key' field in 'LabelSelectorRequirement'
	appLabelSelectorKey = "app"
	// DefaultNodeAffinity is the NodeAffinity to be used when labelSelector is nil
	DefaultNodeAffinity = &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						corev1.NodeSelectorRequirement{
							Key:      NodeAffinityKey,
							Operator: corev1.NodeSelectorOpExists,
						},
					},
				},
			},
		},
	}
	// DaemonPlacements map contains the default placement configs for the
	// various OCS daemons
	DaemonPlacements = map[string]rook.Placement{
		"all": rook.Placement{
			Tolerations: []corev1.Toleration{
				corev1.Toleration{
					Key:      NodeTolerationKey,
					Operator: corev1.TolerationOpEqual,
					Value:    "true",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		},

		"mon": rook.Placement{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					corev1.WeightedPodAffinityTerm{
						Weight: 100,
						PodAffinityTerm: corev1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									metav1.LabelSelectorRequirement{
										Key:      appLabelSelectorKey,
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{"rook-ceph-mon"},
									},
								},
							},
							TopologyKey: corev1.LabelHostname,
						},
					},
				},
			},
		},

		"osd": rook.Placement{
			Tolerations: []corev1.Toleration{
				corev1.Toleration{
					Key:      NodeTolerationKey,
					Operator: corev1.TolerationOpEqual,
					Value:    "true",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			PodAntiAffinity: &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					corev1.WeightedPodAffinityTerm{
						Weight: 100,
						PodAffinityTerm: corev1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									metav1.LabelSelectorRequirement{
										Key:      appLabelSelectorKey,
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{"rook-ceph-osd"},
									},
								},
							},
							TopologyKey: corev1.LabelHostname,
						},
					},
				},
			},
		},

		"rgw": rook.Placement{
			Tolerations: []corev1.Toleration{
				corev1.Toleration{
					Key:      NodeTolerationKey,
					Operator: corev1.TolerationOpEqual,
					Value:    "true",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								metav1.LabelSelectorRequirement{
									Key:      appLabelSelectorKey,
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"rook-ceph-rgw"},
								},
							},
						},
						TopologyKey: corev1.LabelHostname,
					},
				},
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					corev1.WeightedPodAffinityTerm{
						Weight: 100,
						PodAffinityTerm: corev1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									metav1.LabelSelectorRequirement{
										Key:      appLabelSelectorKey,
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{"rook-ceph-rgw"},
									},
								},
							},
							TopologyKey: corev1.LabelZoneFailureDomain,
						},
					},
				},
			},
		},

		"mds": rook.Placement{
			Tolerations: []corev1.Toleration{
				corev1.Toleration{
					Key:      NodeTolerationKey,
					Operator: corev1.TolerationOpEqual,
					Value:    "true",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								metav1.LabelSelectorRequirement{
									Key:      appLabelSelectorKey,
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"rook-ceph-mds"},
								},
							},
						},
						TopologyKey: corev1.LabelHostname,
					},
				},
			},
		},

		"noobaa-core": rook.Placement{
			Tolerations: []corev1.Toleration{
				corev1.Toleration{
					Key:      NodeTolerationKey,
					Operator: corev1.TolerationOpEqual,
					Value:    "true",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		},
	}
)
