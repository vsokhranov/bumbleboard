podTemplate(yaml: '''
              kind: Pod
              spec:
                containers:
                - name: kaniko
                  image: gcr.io/kaniko-project/executor:latest
                  imagePullPolicy: Always
                  command:
                  - sleep
                  args:
                  - 99d
'''
  ) {

  node(POD_LABEL) {
    stage('Build with Kaniko') {
      git 'https://github.com/weirdvic/bumbleboard.git'
      container('kaniko') {
        sh '/kaniko/executor -f `pwd`/Dockerfile -c `pwd` --insecure --skip-tls-verify --cache=true --destination=registry.devops-tools.svc.cluster.local/bumbleboard/bumbleboard:latest'
      }
    }
  }
}
