package powervs

import (
	"fmt"
	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/installconfig"
	"github.com/openshift/installer/pkg/asset/manifests/capiutils"
	"github.com/openshift/installer/pkg/types/powervs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capibm "sigs.k8s.io/cluster-api-provider-ibmcloud/api/v1beta2"
	"time"
)

// GenerateClusterAssets generates the manifests for the cluster-api.
func GenerateClusterAssets(installConfig *installconfig.InstallConfig, clusterID *installconfig.ClusterID, bucket, object string) (*capiutils.GenerateClusterAssetsOutput, error) {
	manifests := []*asset.RuntimeFile{}
	vpcRegion := powervs.Regions[installConfig.Config.Platform.PowerVS.Region].VPCRegion
	imageImport := &capibm.IBMPowerVSImage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IBMPowerVSImage",
			APIVersion: capibm.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "capi-powervs-karthik-image",
			Namespace: capiutils.Namespace,
		},
		Spec: capibm.IBMPowerVSImageSpec{
			ClusterName:  "capi-powervs-karthik",
			Bucket:       &bucket,
			Object:       &object,
			Region:       &vpcRegion,
			DeletePolicy: "retain",
		},
	}

	manifests = append(manifests, &asset.RuntimeFile{
		Object: imageImport,
		File:   asset.File{Filename: "00_powervs-image_import.yaml"},
	})

	powerVSCluster := &capibm.IBMPowerVSCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IBMPowerVSCluster",
			APIVersion: capibm.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "capi-powervs-karthik",
			Namespace: capiutils.Namespace,
		},
		Spec: capibm.IBMPowerVSClusterSpec{
			Zone:          &installConfig.Config.Platform.PowerVS.Zone,
			ResourceGroup: &installConfig.Config.Platform.PowerVS.PowerVSResourceGroup,
			CosBucket: &capibm.CosBucket{
				Name:                 fmt.Sprintf("openshift-bootstrap-data-%s", clusterID.InfraID),
				PresignedURLDuration: &metav1.Duration{Duration: 1 * time.Hour},
			},
			ControlPlaneLoadBalancer: &capibm.VPCLoadBalancerSpec{
				Name: "capi-powervs-karthik-loadbalancer",
				AdditionalListeners: []capibm.AdditionalListenerSpec{
					{
						Port:     22623,
						Protocol: "TCP",
					},
				},
			},
		},
	}
	//
	//Spec: capibm.IBMPowerVSClusterSpec{
	//	Network: capibm.IBMPowerVSResourceReference{
	//		Name: pointer.String("DHCPSERVERcapi-powervs-karthik-dhcp_Private"),
	//	},
	//	ServiceInstance: &capibm.IBMPowerVSResourceReference{
	//		Name: pointer.String("capi-powervs-karthik-serviceInstance"),
	//	},
	//	Zone:          &installConfig.Config.Platform.PowerVS.Zone,
	//	ResourceGroup: &installConfig.Config.Platform.PowerVS.PowerVSResourceGroup,
	//	VPC: &capibm.VPCResourceReference{
	//		ID: pointer.String("r050-012116cd-611c-4caa-a9d2-f8fbbdfcfde1"),
	//	},
	//	VPCSubnet: &capibm.Subnet{
	//		Name: pointer.String("capi-powervs-karthik-vpcsubnet"),
	//	},
	//	TransitGateway: &capibm.TransitGateway{
	//		Name: pointer.String("capi-powervs-karthik-transitgateway"),
	//	},
	//	CosBucket: &capibm.CosBucket{
	//		Name:                 fmt.Sprintf("openshift-bootstrap-data-%s", clusterID.InfraID),
	//		PresignedURLDuration: &metav1.Duration{Duration: 1 * time.Hour},
	//	},
	//},
	manifests = append(manifests, &asset.RuntimeFile{
		Object: powerVSCluster,
		File:   asset.File{Filename: "01_powervs-cluster.yaml"},
	})

	return &capiutils.GenerateClusterAssetsOutput{
		Manifests: manifests,
		InfrastructureRef: &corev1.ObjectReference{
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta2",
			Kind:       "IBMPowerVSCluster",
			Name:       powerVSCluster.Name,
			Namespace:  powerVSCluster.Namespace,
		},
	}, nil
}
