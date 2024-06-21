// The library version is controlled from the Jenkins configuration
// To force a version add after lib '@' followed by the version.
@Library(value = 'msaas-shared-lib', changelog = false) _

node {
    // setup the global static configuration
    config = setupMsaasPipeline('msaas-config.yaml')
}

pipeline {

    options {
        preserveStashes(buildCount: 5)
    }

    agent {
        kubernetes {
            label "${config.pod_label}"
            yaml "${config.KubernetesPods}"
        }
    }

    post {
        always {
            sendMetrics(config)
        }
        fixed {
            emailext(
                    subject: "Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' ${currentBuild.result}",
                    body: """
                        Job Status for '${env.JOB_NAME} [${env.BUILD_NUMBER}]': ${currentBuild.result}\n\nCheck console output at ${env.BUILD_URL}
                """,
                    to: 'some_email@intuit.com'
            )
        }
        unsuccessful {
            emailext(
                    subject: "Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' ${currentBuild.result}",
                    body: """
                        Job Status for '${env.JOB_NAME} [${env.BUILD_NUMBER}]': ${currentBuild.result}\n\nCheck console output at ${env.BUILD_URL}
                """,
                    to: 'some_email@intuit.com'
            )
        }
    }

    stages {
        stage('PRE-BUILD:') {
            when {
                anyOf {
                    branch 'argocd'
                    buildingTag()
                    changeRequest()
                }
            }
            steps {
                setupCapabilities(config, [filepathExclusionRegexes: []])
            }
        }
        stage('BUILD:') {
            when {
                anyOf {
                    branch 'argocd'
                    buildingTag()
                    changeRequest()
                }
            }
            stages {
                stage('Docker Multi Stage Build') {
                    steps {
                        container('podman') {
                            withProdReadOnlyGitCredentials(config) {
                                script {
                                    // Get version from tag, default to empty string if not building from tag
                                    def version = env.TAG_NAME ?: ''

                                    // Build the image
                                    podmanBuild("--build-arg=VERSION=${version} --rm=false -t ${config.image_full_name} .")

                                    // Extract the code coverage report from the container
                                    sh label: 'podman create', script: "podman create --name coverage ${config.image_full_name}"
                                    sh label: 'podman cp', script: "podman cp coverage:/coverage.out coverage.out"

                                    // Publish the image
                                    podmanPush(config)
                                }
                            }
                        }
                    }
                }
                stage('Publish') {
                    parallel {
                        stage('Report Coverage') {
                            steps {
                                container('jnlp') {
                                    codeCov(config)
                                }
                            }
                        }
                        stage('CPD Certification & Publish') {
                            steps {
                                container('cpd2') {
                                    intuitCPD2Podman(config, "-i ${config.image_full_name} --buildfile Dockerfile")
                                }
                                container('podman') {
                                    podmanPull(config, config.image_full_name)
                                    podmanInspect(config, '-s', 'image-metadata.json')
                                    archiveArtifacts(artifacts: 'image-metadata.json', allowEmptyArchive: true)
                                }
                            }
                        }
                        stage('Code Analysis') {
                            when { expression { return config.SonarQubeAnalysis } }
                            steps {
                                container('podman') {
                                    script {
                                        // copy from container bundle for sonar analysis
                                        String image = podmanFindImage([image: 'build', build: env.BUILD_URL])
                                        podmanMount(image, { steps, mount ->
                                            steps.sh(label: 'copy outputs to workspace', script: "cp -r ${mount}/usr/src ${env.WORKSPACE}/bundle")
                                        })
                                        podmanBuild("-f Dockerfile.sonar --build-arg=\"sonar=${config.SonarQubeEnforce}\" .")
                                    }
                                }
                            }
                        }
                        stage('Render Manifests') {
                            when {
                                beforeOptions true
                                allOf {
                                    branch 'argocd'
                                    buildingTag()
                                    not { changeRequest() }
                                }
                            }
                            steps {
                                container('cdtools') {
                                    gitOpsRenderAllManifests(config)
                                }
                            }
                        }
                    }
                }
                // jira transitioning
                stage('Transition Jira Tickets') {
                    steps {
                        script {
                            if (env.BRANCH_NAME != 'argocd' && changeRequest()) {
                                transitionJiraTickets(config, 'Ready for Review')
                            } else if (env.BRANCH_NAME == 'argocd') {
                                transitionJiraTickets(config, 'Closed')
                            }
                        }
                    }
                }
            }
        }

    }
}
