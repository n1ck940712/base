pipeline {
    agent any

    environment {
        DOCKER_REGISTRY_PATH='docker-registry.r4espt.com/r4pid/'
    }

    stages {
        stage('Build') {
            steps {
                script {
                    if (env.GIT_BRANCH == 'origin/master') {
                        env.IMAGE_TYPE = 'release'
                    } else {
                        env.IMAGE_TYPE = 'dev'
                    }

                    env.SLACK_PREFIX = env.JOB_NAME + ' #' + env.BUILD_NUMBER + ': '
                    env.CHANGES_TEXT_FILE_PATH = 'build-' + env.BUILD_NUMBER + '-changes.txt'
                }
                echo "Building branch $GIT_BRANCH"
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Started')
                sh 'git log $GIT_PREVIOUS_SUCCESSFUL_COMMIT..$GIT_COMMIT --oneline --abbrev-commit --pretty=format:"%h [%ad (%cr)] %s - %an" >  $CHANGES_TEXT_FILE_PATH'
                /*Build Mini Game Backend Golang api */
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Building minigame-backend-golang-api ' + env.IMAGE_TYPE + ' docker image ')
                sh './build.sh -i api -b $IMAGE_TYPE -v "$BUILD_NUMBER" -r "$DOCKER_REGISTRY_PATH"'
                script {
                    env.API_IMAGE = sh(script: 'cat .last_built_image', returnStdout: true).trim()
                }
                slackUploadFile filePath: env.CHANGES_TEXT_FILE_PATH, initialComment: env.SLACK_PREFIX + 'minigame-backend-golang-api built with tag ' + env.API_IMAGE
                /*Build Mini Game Backend Golang websocket */
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Building minigame-backend-golang-websocket ' + env.IMAGE_TYPE + ' docker image ')
                sh './build.sh -i websocket -b $IMAGE_TYPE -v "$BUILD_NUMBER" -r "$DOCKER_REGISTRY_PATH"'
                script {
                    env.WEBSOCKET_IMAGE = sh(script: 'cat .last_built_image', returnStdout: true).trim()
                }
                slackUploadFile filePath: env.CHANGES_TEXT_FILE_PATH, initialComment: env.SLACK_PREFIX + 'minigame-backend-golang-websocket built with tag ' + env.WEBSOCKET_IMAGE
                /*Build Mini Game Backend Golang Gameloop LOL Tower */
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Building minigame-backend-golang-gameloop-lol-tower ' + env.IMAGE_TYPE + ' docker image ')
                sh './build.sh -i gameloop-lol-tower -b $IMAGE_TYPE -v "$BUILD_NUMBER" -r "$DOCKER_REGISTRY_PATH"'
                script {
                    env.GAMELOOP_LOL_TOWER_IMAGE = sh(script: 'cat .last_built_image', returnStdout: true).trim()
                }
                slackUploadFile filePath: env.CHANGES_TEXT_FILE_PATH, initialComment: env.SLACK_PREFIX + 'minigame-backend-golang-gameloop-lol-tower built with tag ' + env.GAMELOOP_LOL_TOWER_IMAGE
                /*Build Mini Game Backend Golang Gameloop LOL Couple */
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Building minigame-backend-golang-gameloop-lol-couple ' + env.IMAGE_TYPE + ' docker image ')
                sh './build.sh -i gameloop-lol-couple -b $IMAGE_TYPE -v "$BUILD_NUMBER" -r "$DOCKER_REGISTRY_PATH"'
                script {
                    env.GAMELOOP_LOL_COUPLE_IMAGE = sh(script: 'cat .last_built_image', returnStdout: true).trim()
                }
                slackUploadFile filePath: env.CHANGES_TEXT_FILE_PATH, initialComment: env.SLACK_PREFIX + 'minigame-backend-golang-gameloop-lol-couple built with tag ' + env.GAMELOOP_LOL_COUPLE_IMAGE
            }
        }
        stage('Test') {
            steps {
                script {
                    // if (env.IMAGE_TYPE != 'release' && env.ESPORTS_DISABLE_UNIT_TEST != '1') {
                    //     slackSend color: 'good', message: env.SLACK_PREFIX + 'Running unit tests'
                    //     sh(script: 'ARGS=$API_IMAGE make test_ci')
                    // }
                    echo "Test"
                }
            }
        }
        stage('Upload') {
            steps {
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Uploading minigame-backend-golang-api ' + env.IMAGE_TYPE + ' docker image ')
                sh 'docker push $API_IMAGE'
                script {
                    if (env.IMAGE_TYPE == 'release') {
                        sh(script: 'docker push $DOCKER_REGISTRY_PATH"minigame-backend-golang-api:latest"')
                    } else {
                        sh(script: 'docker push $DOCKER_REGISTRY_PATH"minigame-backend-golang-api:dev"')
                    }
                }
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Uploading minigame-backend-golang-websocket ' + env.IMAGE_TYPE + ' docker image ')
                sh 'docker push $WEBSOCKET_IMAGE'
                script {
                    if (env.IMAGE_TYPE == 'release') {
                        sh(script: 'docker push $DOCKER_REGISTRY_PATH"minigame-backend-golang-websocket:latest"')
                    } else {
                        sh(script: 'docker push $DOCKER_REGISTRY_PATH"minigame-backend-golang-websocket:dev"')
                    }
                }
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Uploading minigame-backend-golang-gameloop-lol-tower ' + env.IMAGE_TYPE + ' docker image ')
                sh 'docker push $GAMELOOP_LOL_TOWER_IMAGE'
                script {
                    if (env.IMAGE_TYPE == 'release') {
                        sh(script: 'docker push $DOCKER_REGISTRY_PATH"minigame-backend-golang-gameloop-lol-tower:latest"')
                    } else {
                        sh(script: 'docker push $DOCKER_REGISTRY_PATH"minigame-backend-golang-gameloop-lol-tower:dev"')
                    }
                }
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Uploading minigame-backend-golang-gameloop-lol-couple ' + env.IMAGE_TYPE + ' docker image ')
                sh 'docker push $GAMELOOP_LOL_COUPLE_IMAGE'
                script {
                    if (env.IMAGE_TYPE == 'release') {
                        sh(script: 'docker push $DOCKER_REGISTRY_PATH"minigame-backend-golang-gameloop-lol-couple:latest"')
                    } else {
                        sh(script: 'docker push $DOCKER_REGISTRY_PATH"minigame-backend-golang-gameloop-lol-couple:dev"')
                    }
                }
            }
        }
        stage('Deploy') {
            steps {
                script {
                    if (env.GIT_BRANCH == 'origin/master') {
                        env.K8S_NAMESPACE = 'live'
                    } else {
                        env.K8S_NAMESPACE = 'dev'
                    }
                }
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Updating images for r4pid-minigame-backend-go-api, r4pid-minigame-backend-go-websocket and r4pid-minigame-backend-go-gameloop-lol-tower')

                sh 'kubectl set env deployment/r4pid-minigame-backend-go-api IMAGE_TAG=minigame-backend-golang-api:$IMAGE_TYPE-$BUILD_NUMBER -n $K8S_NAMESPACE'
                sh 'kubectl set image deployment/r4pid-minigame-backend-go-api r4pid-minigame-backend-go-api=$API_IMAGE -n $K8S_NAMESPACE --record'
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Waiting for r4pid-minigame-backend-go-api replicas to be ready')
                sh 'sh ./build/wait-k8s-deployment.sh -d "r4pid-minigame-backend-go-api" -n $K8S_NAMESPACE'

                sh 'kubectl set env deployment/r4pid-minigame-backend-go-gameloop-lol-tower IMAGE_TAG=minigame-backend-golang-gameloop-lol-tower:$IMAGE_TYPE-$BUILD_NUMBER -n $K8S_NAMESPACE'
                sh 'kubectl set image deployment/r4pid-minigame-backend-go-gameloop-lol-tower r4pid-minigame-backend-go-gameloop-lol-tower=$GAMELOOP_LOL_TOWER_IMAGE -n $K8S_NAMESPACE --record'
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Waiting for r4pid-minigame-backend-go-gameloop-lol-tower replicas to be ready')
                sh 'sh ./build/wait-k8s-deployment.sh -d "r4pid-minigame-backend-go-gameloop-lol-tower" -n $K8S_NAMESPACE'

                sh 'kubectl set env deployment/r4pid-minigame-backend-go-gameloop-lol-couple IMAGE_TAG=minigame-backend-golang-gameloop-lol-couple:$IMAGE_TYPE-$BUILD_NUMBER -n $K8S_NAMESPACE'
                sh 'kubectl set image deployment/r4pid-minigame-backend-go-gameloop-lol-couple r4pid-minigame-backend-go-gameloop-lol-couple=$GAMELOOP_LOL_COUPLE_IMAGE -n $K8S_NAMESPACE --record'
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Waiting for r4pid-minigame-backend-go-gameloop-lol-couple replicas to be ready')
                sh 'sh ./build/wait-k8s-deployment.sh -d "r4pid-minigame-backend-go-gameloop-lol-couple" -n $K8S_NAMESPACE'

                sh 'kubectl set env deployment/r4pid-minigame-backend-go-websocket IMAGE_TAG=minigame-backend-golang-websocket:$IMAGE_TYPE-$BUILD_NUMBER -n $K8S_NAMESPACE'
                sh 'kubectl set image deployment/r4pid-minigame-backend-go-websocket r4pid-minigame-backend-go-websocket=$WEBSOCKET_IMAGE -n $K8S_NAMESPACE --record'
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Waiting for r4pid-minigame-backend-go-websocket replicas to be ready')
                sh 'sh ./build/wait-k8s-deployment.sh -d "r4pid-minigame-backend-go-websocket" -n $K8S_NAMESPACE'

            }
        }
        stage('Cleanup') {
            steps {
                slackSend (color: 'good', message: env.SLACK_PREFIX + 'Cleaning up built minigame-backend-golang-api minigame-backend-golang-websocket minigame-backend-golang-gameloop-lol-couple & minigame-backend-golang-gameloop-lol-tower images ')
                echo "Cleaning up test push."
                sh 'docker rmi $API_IMAGE'
                sh 'docker rmi $GAMELOOP_LOL_TOWER_IMAGE'
                sh 'docker rmi $GAMELOOP_LOL_COUPLE_IMAGE'
                sh 'docker rmi $WEBSOCKET_IMAGE'

                script {
                    if (env.IMAGE_TYPE == 'release') {
                        sh(script: 'docker rmi $DOCKER_REGISTRY_PATH"minigame-backend-golang-api:latest"')
                        sh(script: 'docker rmi $DOCKER_REGISTRY_PATH"minigame-backend-golang-gameloop-lol-tower:latest"')
                        sh(script: 'docker rmi $DOCKER_REGISTRY_PATH"minigame-backend-golang-gameloop-lol-couple:latest"')
                        sh(script: 'docker rmi $DOCKER_REGISTRY_PATH"minigame-backend-golang-websocket:latest"')
                    } else {
                        sh(script: 'docker rmi $DOCKER_REGISTRY_PATH"minigame-backend-golang-api:dev"')
                        sh(script: 'docker rmi $DOCKER_REGISTRY_PATH"minigame-backend-golang-gameloop-lol-tower:dev"')
                        sh(script: 'docker rmi $DOCKER_REGISTRY_PATH"minigame-backend-golang-gameloop-lol-couple:dev"')
                        sh(script: 'docker rmi $DOCKER_REGISTRY_PATH"minigame-backend-golang-websocket:dev"')
                    }
                }
            }
        }
    }
    post {
        success {
            slackSend (color: 'good', message: env.SLACK_PREFIX + 'Done')
        }
        failure {
            slackSend (color: 'error', message: env.SLACK_PREFIX + 'Failed')
        }
    }
}