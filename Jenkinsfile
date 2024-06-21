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
                    branch 'master'
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
                    branch 'master'
                    changeRequest()
                }
            }
            stages {
                stage('Docker Multi Stage Build') {
                    steps {
                        container('podman') {
                            withProdReadOnlyGitCredentials(config) {
                                podmanBuild("--rm=false --target=build -t ${config.image_full_name}_test .")
                                podmanRun("${config.image_full_name}_test", "--name=${config.git_org}_${config.service_name}_unit_test_${env.BUILD_NUMBER}")
                                sh label: 'docker cp', script: "podman cp ${config.git_org}_${config.service_name}_unit_test_${env.BUILD_NUMBER}:/go/src/github.intuit.com/${config.git_org}/${config.service_name}/coverage.out coverage.out"
                                podmanBuild("--rm=false -t ${config.image_full_name} .")
                                podmanPush(config)
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
                            when {expression {return config.SonarQubeAnalysis}}
                            steps {
                                container('podman') {
                                    script {
                                        // copy from container bundle for sonar analysis
                                        String image = podmanFindImage([image: 'build', build: env.BUILD_URL])
                                        podmanMount(image, {steps,mount ->
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
                                    branch 'master'
                                    not {changeRequest()}
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
                            if (env.BRANCH_NAME != 'master' && changeRequest()) {
                                transitionJiraTickets(config, 'Ready for Review')
                            } else if (env.BRANCH_NAME == 'master') {
                                transitionJiraTickets(config, 'Closed')
                            }
                        }
                    }
                }
            }
        }
        stage('qal-usw2-eks') {
            when {
                beforeOptions true
                allOf {
                    branch 'master'
                    not {changeRequest()}
                }
            }
            options {
                lock(resource: getEnv(config, 'qal-usw2-eks').namespace, inversePrecedence: true)
                timeout(time: 32, unit: 'MINUTES')
            }
            stages {
                stage('Scorecard Check') {
                    when {expression {return config.enableScorecardReadinessCheck}}
                    steps {
                        scorecardPreprodReadiness(config, 'qal-usw2-eks')
                    }
                }
                stage('Deploy') {
                    steps {
                        container('cdtools') {
                            // This has to be the first action in the first sub-stage
                            milestone(ordinal: 10, label: 'Deploy-qal-usw2-eks-milestone')
                            gitOpsDeploy(config, 'qal-usw2-eks', config.image_full_name)
                        }
                    }
                }
                stage('Transition Jira Tickets') {
                    when {expression {return config.enableJiraTransition}}
                    steps {
                        transitionJiraTickets(config, 'Deployed to PreProd')
                    }
                }
            }
        }
        stage('e2e-usw2-eks') {
            when {
                beforeOptions true
                allOf {
                    branch 'master'
                    not {changeRequest()}
                }
            }
            options {
                lock(resource: getEnv(config, 'e2e-usw2-eks').namespace, inversePrecedence: true)
                timeout(time: 32, unit: 'MINUTES')
            }
            stages {
                stage('Scorecard Check') {
                    when {expression {return config.enableScorecardReadinessCheck}}
                    steps {
                        scorecardPreprodReadiness(config, 'e2e-usw2-eks')
                    }
                }
                stage('Deploy') {
                    steps {
                        container('cdtools') {
                            // This has to be the first action in the first sub-stage
                            milestone(ordinal: 20, label: 'Deploy-e2e-usw2-eks-milestone')
                            gitOpsDeploy(config, 'e2e-usw2-eks', config.image_full_name)
                        }
                    }
                }
                stage('Transition Jira Tickets') {
                    when {expression {return config.enableJiraTransition}}
                    steps {
                        transitionJiraTickets(config, 'Deployed to PreProd')
                    }
                }
            }
        }
    }
}