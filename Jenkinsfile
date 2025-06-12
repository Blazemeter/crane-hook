@Library('jenkins_library')
import com.blazemeter.jenkins.lib.DockerTag

IMAGE_NAME = 'us.gcr.io/verdant-bulwark-278/crane-hook'

pipeline {
    agent {
        docker {
            image 'us.gcr.io/verdant-bulwark-278/jenkins-docker-agent:node18.latest'
            args '--network host -u root -v /home/jenkins/tools/:/home/jenkins/tools/ -v /var/run/docker.sock:/var/run/docker.sock'
            label 'docker'
        }
    }

    parameters {
        string(name: 'BRANCH_NAME_PARAM', defaultValue: 'main', description: '')
    }

    environment {
        SENDER = 'jenkins@blazemeter.com'
    }
    
    stages {
        stage('init') {
            steps {
                script {
                    initJenkinsGlobal()
                    String buildDate = new Date().format("yyyyMMddHHmmss", TimeZone.getTimeZone('UTC'))
                    //currentBuild.displayName = "#${buildDate}"
                    env.BUILD_TIMESTAMP_ID = buildDate
                    def branch = env.BRANCH_NAME ?: params.BRANCH_NAME_PARAM ?: 'main'
                    //env.UNIFIED_BRANCH_NAME = env.BRANCH_NAME.replaceAll("/", "_")
                    env.TAG = (env.BRANCH_NAME == "main") ? "latest" : env.BRANCH_NAME
                }
            }
        }

        stage('Build Docker Image') {
            steps {
                sh "docker build -t ${IMAGE_NAME} -f Dockerfile ."
                script {
                    def tags = new DockerTag()
                    tags.addTag("${env.TAG}")
                    tags.addTag("${env.BRANCH_NAME}-${env.BUILD_NUMBER}")

                    pushImageToAllRegistries("${IMAGE_NAME}", "${IMAGE_NAME}", tags)
                }
            }
        }

        stage('Perform WhiteSource scan') {
            when {
                environment name: 'BRANCH_NAME', value: 'main'
            }
            steps {
                script {
                    def projectName = "${IMAGE_NAME}"
                    whiteSourceScan("${IMAGE_NAME}", params.BRANCH_NAME_PARAM)
                }
            }
        }
    }

    post {
        cleanup {
            cleanWs()
        }
        failure {
            // Send an email if the build fails
            notifyJobFailureEmailToAuthor(sender: "${env.SENDER}")
        }
    }
}
