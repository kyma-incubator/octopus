# Docs

## Overview

The folder contains documents that provide an insight into Prow configuration, development, and testing.

<!-- Update the list each time you modify the document structure in this folder. -->

Read the documents to learn how to:

- [Configure Prow on a production cluster](./production-cluster-configuration.md) based on the preconfigured Google Cloud Storage (GCS) resources.
- [Create a service account](./prow-secrets-management.md) and store its encrypted key in a GCS bucket.
- [Install and configure Prow](./prow-installation-on-forks.md) on a forked repository to test and develop it on your own.
- [Install and manage monitoring](./prow-monitoring.md) on a Prow cluster.
- [Create, modify, and remove standard component jobs](./component-jobs.md) for the Prow pipeline.

Find out more about:

- [Prow architecture](./prow-architecture.md) and its setup in the Kyma project.
- [ProwJobs](./prow-jobs.md) for details on ProwJobs.
- [Obligatory security measures](obligatory-security-measures.md) to take regularly for the Prow production cluster and when someone leaves the Kyma project.
- [Presets](./presets.md) you can use to define ProwJobs.
- [Authorization](./authorization.md) concepts employed in Prow.