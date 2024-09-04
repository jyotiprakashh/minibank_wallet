pipeline {
    agent any

    stages {
        stage('Checkout') {
            steps {
                git branch: 'main', url: 'https://github.com/jyotiprakashh/minibank_wallet.git'
            }
        }
        stage('Build') {
            steps {
                sh 'go build -v ./...'
            }
        }
        stage('Test') {
            steps {
                sh 'go test -v ./...'
            }
        }
        stage('Terraform Init') {
            steps {
                sh 'terraform init'
            }
        }
        stage('Terraform Apply') {
            steps {
                sh 'terraform apply -auto-approve'
            }
        }
    }
}
