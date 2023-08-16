podTemplate(yaml: '''
              kind: Pod
              spec:
                containers:
                - name: kaniko
                  image: gcr.io/kaniko-project/executor:latest
                  imagePullPolicy: Always
                args:
                  - "--dockerfile=Dockerfile"
                  - "--context=https://github.com/weirdvic/bumbleboard.git"
                  - "--destination=registry.devops-tools.svc.cluster.local/bumbleboard/bumbleboard:latest>"
'''
  )
