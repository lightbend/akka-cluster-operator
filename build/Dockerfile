FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

### Required OpenShift Certified Labels
LABEL name="Akka Cluster Operator" \
      vendor="Lightbend, Inc." \
      version="v0.2.2" \
      release="0.2.2" \
      summary="Kubernetes Operator for Akka Cluster applications." \
      description="Manage applications designed for [Akka Cluster](https://doc.akka.io/docs/akka/current/common/cluster.html) and the Lightbend Platform."

# Required to be in /licenses
RUN  mkdir /licenses
COPY LICENSE /licenses

ENV OPERATOR=/usr/local/bin/akka-cluster-operator \
    USER_UID=1001 \
    USER_NAME=akka-cluster-operator

# install operator binary
COPY build/_output/bin/akka-cluster-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
