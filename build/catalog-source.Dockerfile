FROM quay.io/operator-framework/upstream-registry-builder as builder
# Add noobaa manifests
COPY build/_output/olm manifests/noobaa
# Add lib-bucket-provisioner manifests
COPY deploy/obc/lib-bucket-provisioner.package.yaml manifests/lib-bucket-provisioner/
COPY deploy/obc/lib-bucket-provisioner.v1.0.0.clusterserviceversion.yaml manifests/lib-bucket-provisioner/1.0.0/
COPY deploy/obc/objectbucket_v1alpha1_objectbucket_crd.yaml manifests/lib-bucket-provisioner/1.0.0/
COPY deploy/obc/objectbucket_v1alpha1_objectbucketclaim_crd.yaml manifests/lib-bucket-provisioner/1.0.0/
RUN ./bin/initializer -o ./bundles.db

FROM scratch
COPY --from=builder /build/bundles.db /bundles.db
COPY --from=builder /build/bin/registry-server /registry-server
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
EXPOSE 50051
ENTRYPOINT ["/registry-server"]
CMD ["--database", "bundles.db"]
