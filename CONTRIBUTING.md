# Contributing to ibp-genai-service

Thank you for your interest in contributing to the ibp-genai-service! This document provides guidelines and instructions for contributing to this project.

## Before PR

### Pre-work Necessary
Before you contribute a feature or bugfix, please:
- File an issue in the tracker to discuss potential changes and gather feedback.
- Engage in design discussions and contribute to architecture decision records (ADRs) if applicable.

### Turnaround Time
Expect a turnaround time of approximately 2-3 days for the review of feature proposals.

## During PR

### Code Review Process
- Your code will be reviewed by the maintainers listed in the [CODEOWNERS file](.github/CODEOWNERS).
- Trusted committers and maintainers are responsible for reviewing and merging PRs.

### SLA for PR Reviews
- The expected turnaround time for PR reviews and new commits to existing PRs is within 48 hours.

### Git Branching Strategy
- Branch names should follow a format that includes the issue number, e.g., `JIRA-12345`.

### Documentation and Changelog
- Ensure all new code functionalities are well-documented in the code itself and in the external documentation.

### Code Style and Testing
- Adhere to the coding standards as documented in the project.
- Include unit tests with new code to maintain and improve the coverage. A testing strategy should be discussed in the PR if it introduces significant changes.

### Design Artifacts
- Include any design artifacts relevant to the PR in the `docs/design` folder.

## After PR

### Post-Merge Process
- After code is merged, it will be built and deployed to the preprod and production environments based on the project pipeline.
- Contributors are expected to monitor their changes in the preprod environment for at least 24 hours.

### Deployment Timing
- Expect to see your changes built and deployed within 24 hours after merge.

### Validation of Changes
- Contributors can validate their changes by checking the deployment status and logs in the preprod environment.

## Contact Information

### Maintainers
- For a list of project maintainers, refer to the [CODEOWNERS file](.github/CODEOWNERS).

### Support
- Slack: [#ibp-community](https://intuit.enterprise.slack.com/archives/C9YFBNJBV)
