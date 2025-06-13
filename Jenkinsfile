@Library('jenkins_library')
import com.blazemeter.jenkins.lib.DockerTag

def REGISTRY = "gcr.io"
def FALLBACK_REGISTRY = "us.gcr.io"
def PROJECT_ID = "verdant-bulwark-278"
def IMAGE_NAME = "cranehook"
def IMAGE_TAG = "${env.TAG}"
def IMAGE_FULL = "${REGISTRY}/${PROJECT_ID}/${IMAGE_NAME}:${IMAGE_TAG}"
def IMAGE_FOR_HELPER = "${FALLBACK_REGISTRY}/${PROJECT_ID}/${IMAGE_NAME}:${IMAGE_TAG}"
def COMPONENT_NAME = "${REGISTRY}/${PROJECT_ID}/${IMAGE_NAME}"

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
                    env.BUILD_TIMESTAMP_ID = buildDate
                    def branchName = env.BRANCH_NAME ?: params.BRANCH_NAME_PARAM ?: 'main'
                    env.BRANCH_NAME = branchName
                    env.TAG = (branchName == "main") ? "latest" : env.BRANCH_NAME
        
                    echo "Set TAG=${env.TAG}"
                }
            }
        }

        stage('Install Go') {
            steps {
                sh '''
                    echo "Installing Go locally..."
                    curl -sSL https://golang.org/dl/go1.21.6.linux-amd64.tar.gz -o go.tar.gz
                    mkdir -p $WORKSPACE/go
                    tar -C $WORKSPACE/go --strip-components=1 -xzf go.tar.gz
                    rm go.tar.gz
                '''
            }
        }

        stage('Build Go Binary') {
            steps {
                sh '''
                    echo "Building Go binary..."
                    export GOROOT=$WORKSPACE/go
                    export PATH=$GOROOT/bin:$PATH
                    export GOOS=linux
                    export GOARCH=amd64
                    go version
                    go build -o cranehook
                '''
            }
        }        

        stage('Build Docker Image') {
            steps {
                sh "docker build -t ${IMAGE_FULL} -f Dockerfile ."
                sh "docker tag ${IMAGE_FULL} ${IMAGE_FOR_HELPER}"
                script {
                    def tags = new DockerTag()
                    tags.addTag("${IMAGE_TAG}")
                    tags.addTag("${env.BRANCH_NAME}-${env.BUILD_NUMBER}")
                    echo "Docker image will be tagged as: ${IMAGE_FOR_HELPER}"
                    pushImageToAllPublicRegistries(IMAGE_FOR_HELPER, IMAGE_NAME, tags)
                }
            }
        }

        stage('Cleanup Binary') {
            steps {
                sh 'rm -f cranehook'
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
