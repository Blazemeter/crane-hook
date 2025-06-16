@Library('jenkins_library')
import com.blazemeter.jenkins.lib.DockerTag

IMAGE_NAME = 'cranehook'

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
                withCredentials([file(credentialsId: 'gcr-pull-secret-json', variable: 'GCP_KEY_FILE')]) {
                script {
                    def REGISTRY = "gcr.io"
                    def PROJECT_ID = "verdant-bulwark-278"
                    def IMAGE_NAME = "cranehook"
                    def TAG = env.TAG ?: "latest"
                    def FULL_IMAGE = "${REGISTRY}/${PROJECT_ID}/${IMAGE_NAME}:${TAG}"

                    // Build the Docker image
                    sh "docker build -t ${FULL_IMAGE} -f Dockerfile ."

                    // Authenticate to gcr.io using service account key
                    sh '''
                        echo "Authenticating to gcr.io..."
                        gcloud auth activate-service-account --key-file=$GCP_KEY_FILE
                        gcloud auth configure-docker gcr.io --quiet
                    '''

                    // Push the image
                    sh "docker push ${FULL_IMAGE}"

                    echo "Docker image pushed to: ${FULL_IMAGE}"
                    }
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