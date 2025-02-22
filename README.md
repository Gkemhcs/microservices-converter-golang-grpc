# Microservices Converter Demo

### **Note:-** 
**⚠️ Google Sign-In can only be displayed on HTTPS domains, not HTTP. For testing and development purposes, we can use it on localhost. Therefore, to use this feature, we access the service on localhost to make it work.**

## Overview

This project demonstrates a microservices architecture implemented in Go (Golang). The system consists of several services that perform various media conversion tasks, such as converting text to speech, video to audio, and images to PDFs. Each service is designed to be independent, scalable, and easy to maintain.

Our website provides a user-friendly interface for performing media conversions. Users can sign in using Google and access converter pages like text-to-speech, video-to-audio, and image-to-pdf. The frontend service, built with the Gin framework, renders HTML templates and connects to backend gRPC services to handle the conversion tasks. The website ensures a seamless experience by integrating various microservices to perform specific conversion functions efficiently and also sign in using google to ease authentication like login and logout.

## Services

| Service Name   | Purpose                                                                 | Stack Used                                                                 | Port |
|----------------|-------------------------------------------------------------------------|----------------------------------------------------------------------------|------|
| Frontend       | Provides a user interface for interacting with the various conversion services. | Gin, Logrus, gRPC, PostgreSQL, OpenTelemetry                               | 8080 |
| File Uploader  | A gRPC service that uploads received files via chunks to Google Cloud Storage and generates signed URLs. | gRPC, Logrus, Google Cloud Storage, OpenTelemetry                          | grpc:8084, metrics-port:9090 |
| Image-to-PDF   | A gRPC service that converts image chunks to PDF and sends the PDF chunks to the File Uploader for a signed URL. | gRPC, Logrus, OpenTelemetry, github.com/signintech/gopdf                   | grpc:8083,metrics-port:9090 |
| Video-to-Audio | A gRPC service that converts video chunks to audio and sends the audio chunks to the File Uploader for a signed URL. | gRPC, Logrus, OpenTelemetry, FFmpeg, github.com/u2takey/ffmpeg-go          | grpc:8082,metrics-port:9090 |
| Text-to-Speech | A gRPC service that converts text to speech and sends the audio chunks to the File Uploader for a signed URL. | gRPC, Logrus, OpenTelemetry, github.com/Duckduckgot/gtts                   | grpc:8081, metrics-port:9090 |
| Postgres | Database to store the generated signed URLs | PostgreSQL | 5432 |
| OpenTelemetry Collector | Otel Collector to collect traces from applications via endpoints, process them, and export them to Zipkin for visualization of traces | OpenTelemetry | **HTTP**: 4318 **gRPC**: 4317 **Zipkin**: 9411 |
| EFK Stack | Logging stack: **Fluentd** for forwarding logs to storage, **Elasticsearch** for storing logs in the form of indices for faster retrieval, and **Kibana** for visualizing logs stored in Elasticsearch | Elasticsearch, Fluentd, Kibana | Kibana: 5601, Elasticsearch: 9200 |

## Project Structure

```
converter-gcp/
├── helm-chart/
├── k8s-templates/
├── protos/
├── src/
│   ├── frontend/
│   ├── text-to-speech/
│   ├── video-to-audio/
│   ├── image-to-pdf/
│   └── file-uploader/
├── values/
├── assets/
├── docker-compose.yaml
└── README.md
```
## Project Architecture
!["Project Architecture"](assets/architecture.jpg)

## Website Images
[Click here for Website related snapshots ](./Info.md)


## Setup and Installation

### Prerequisites

- Docker
- Docker Compose
- Kubernetes (for K8s deployment)
- Helm (for K8s deployment)
- Google Cloud Service Account Key (JSON file)
- Google Cloud Platform (GCP) account

### Setting up Prerequisites 
##### If you are deploying through **Docker** make sure you have Docker installed on your machine. If you are deploying onto Kubernetes, make sure you have kubectl and Helm installed on your machine and also have a Kubernetes cluster (e.g., Kind, Minikube, Managed Kubernetes Clusters).
1. **Get into Google Cloud Console**:
   - Navigate to the [Google Cloud Console](https://console.cloud.google.com/).

2. **Create OAuth Token**:
   - Go to the **Credentials** page under **APIs & Services**.
   - Create an OAuth token of type **Web Application**.
   - Add the following URLs under **Authorized JavaScript origins**:
     ```
     http://localhost
     http://localhost:8080
     ```
   - Add the following URL under **Authorized redirect URIs**:
     ```
     http://localhost:8080/user/google/auth/callback
     ```
   - Make a note of the **Google Client ID** and **Google Client Secret** for use in further sections.
    ![sample config](assets/google-signin-console-section.png)
    *Sample Screenshot of configuration*
3. **Create GCP Bucket and Service Account Key**:
   - Enter your project id with billing account linked
      ```sh
      echo "ENTER YOUR GOOGLE PROJECT ID"
      read PROJECT_ID
      ```
   - Create a GCP bucket for storing files:
     ```sh
     export BUCKET_NAME="microservices-converter-bucket-${PROJECT_ID}"
     gsutil mb gs://$BUCKET_NAME
     ```
   - Create a service account key with **Storage Admin** permissions:
     ```sh
     export GCP_SA="gcs-uploader@${PROJECT_ID}.iam.gserviceaccount.com"
     gcloud iam service-accounts create gcs-uploader --display-name "Service Account for Storage Admin"
     gcloud projects add-iam-policy-binding $PROJECT_ID --member "serviceAccount:$GCP_SA" --role "roles/storage.admin"
     gcloud iam service-accounts keys create key.json --iam-account $GCP_SA
     ```
   - Copy the Downloaded file key.json to project directory

### Docker Deployment 

1. **Clone the repository**:
    ```sh
    git clone https://github.com/Gkemhcs/microservices-converter-golang-grpc.git
    cd microservices-converter-golang-grpc
    ```

2. **Set up environment variables**:
    Ensure you have a `key.json` file for Google Cloud credentials and place it in the project root directory.

3. **Replace placeholders in Docker Compose file**:
    - Open the `docker-compose.yaml` file.
    - Replace `GOOGLE_CLIENT_ID_CREDS` with your Google Client ID.
    - Replace `GOOGLE_CLIENT_SECRET_CREDS` with your Google Client Secret.
    - Replace `GCP_SERVICE_ACCOUNT_NAME` with your GCP Service Account.
    - Replace `GCS_BUCKET` with your GCP Bucket name.

4. **Build and run the services**:
    ```sh
    docker-compose up --build -d

    # For enabling  logging stack run the below command using logging profile
    docker compose --profile logging up -d

    # For enabling monitoring stack in our application run the below command to enable monitoring profile
    docker compose --profile monitoring up -d
    ```

5. **Service URLs and Ports**:
    - Frontend: `http://localhost:8080`
    - Text-to-Speech: `http://localhost:8081`
    - Video-to-Audio: `http://localhost:8082`
    - Image-to-PDF: `http://localhost:8083`
    - File Uploader: `http://localhost:8084`
    - Kibana: `http://localhost:5601`
    - Grafana: `http://localhost:3000`
    - Zipkin: `http://localhost:9411`
    - Prometheus: `http://localhost:9090`

### Kubernetes Deployment

1. **Clone the repository**:
    ```sh
    git clone https://github.com/yourusername/converter-gcp.git
    cd converter-gcp
    ```

2. **Set up environment variables**:
    Ensure you have a `key.json` file downloaded earlier for Google Cloud credentials and place it in the project root directory.

3. **Replace placeholders in Kubernetes values files**:
    - Open the `values/frontend.yaml` file.
    - Replace `GOOGLE_CLIENT_ID_CREDS` with your Google Client ID.
    - Replace `GOOGLE_CLIENT_SECRET_CREDS` with your Google Client Secret.
    - Open the `values/file-uploader.yaml` file.
    - Replace `GCP_SERVICE_ACCOUNT_NAME` with your GCP Service Account.
    - Replace `GCS_BUCKET` with your GCP Bucket name.

4. **Install dependencies**:
    - Deploy PostgreSQL and Redis:
      ```sh
      kubectl apply -f k8s-templates/databases-deployment.yaml
      ```
    - Before deploying OpenTelemetry operator manifests, first deploy the Cert-Manager and OpenTelemetry Operators:
      ```sh
        # install cert-manager
        helm repo add jetstack https://charts.jetstack.io
        helm repo update
        helm install \
          --create-namespace \
          --namespace cert-manager \
          --set installCRDs=true \
          --set global.leaderElection.namespace=cert-manager \
          --set extraArgs={--issuer-ambient-credentials=true} \
          cert-manager jetstack/cert-manager
        # install opentelemetry operator
        kubectl apply -f https://github.com/open-telemetry/opentelemetry-operator/releases/latest/download/opentelemetry-operator.yaml
      ``` 
    - Deploy OpenTelemetry Collector and Zipkin:
      ```sh
      kubectl apply -f k8s-templates/observability.yaml
      ```

5. **Install the EFK stack**:
    - Deploy Elasticsearch:
      ```sh
      helm repo add elastic https://helm.elastic.co
      helm install elasticsearch \
      --set replicas=1 \
      --set persistence.labels.enabled=true elastic/elasticsearch -n logging \
      --create-namespace
      ```
    - Retrieve Elasticsearch Username & Password:
      ```sh
        # for username
        kubectl get secrets --namespace=logging elasticsearch-master-credentials -ojsonpath='{.data.username}' | base64 -d
        # for password
        kubectl get secrets --namespace=logging elasticsearch-master-credentials -ojsonpath='{.data.password}' | base64 -d
      ```
    
    - Deploy Kibana:
      ```sh
        helm install kibana --set service.type=LoadBalancer elastic/kibana -n logging
      ```

    - Deploy Fluentd: 👉 **Note**: Please update the `HTTP_Passwd` field in the `fluentbit-values.yml` file with the password retrieved earlier in step 6: (i.e., NJyO47UqeYBsoaEU)
      ```sh
      helm repo add fluent https://fluent.github.io/helm-charts
      helm install fluent-bit fluent/fluent-bit -f values/fluentbit-values.yaml -n logging
      ```
6. **Install **ISTIO SERVICE MESH** into cluster in sidecar mode**
- First install istioctl to operate with istio

  ```sh
  curl -sL https://istio.io/downloadIstioctl | sh -
  export PATH=$HOME/.istioctl/bin:$PATH
  ```
- Install **Istio** resources into cluster 

    ```sh
    istioctl install --set meshConfig.accessLogFile=/dev/stdout

    # deploy prometheus integration in istio-system namespaces 
    kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.24/samples/addons/prometheus.yaml

    # deploy kiali  integration in istio-system namespace to visualise the mesh
    kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.24/samples/addons/kiali.yaml
    ```

7. **Create Application Namespaces and add the namespaces to mesh**
    ```sh
    kubectl create ns frontend-ns 
    kubectl create ns text-to-speech-ns
    kubectl create ns video-to-audio-ns
    kubectl create ns image-to-pdf-ns
    kubectl create ns file-uploader-ns

    #Add labels to namespace to add namespaces to istio sidecar mesh
    kubectl label namespace frontend-ns istio-injection=enabled
    kubectl label namespace text-to-speech-ns istio-injection=enabled
    kubectl label namespace video-to-audio-ns istio-injection=enabled
    kubectl label namespace image-to-pdf-ns istio-injection=enabled
    kubectl label namespace file-uploader-ns istio-injection=enabled
    ```
8. **Deploy the services using Helm**:
    ```sh
    helm install frontend ./helm-chart/ -f values/frontend.yaml --namespace frontend-ns 
    helm install text-to-speech ./helm-chart/ -f values/text-to-speech.yaml --namespace text-to-speech-ns 
    helm install video-to-audio ./helm-chart/ -f values/video-to-audio.yaml --namespace video-to-audio-ns 
    helm install image-to-pdf ./helm-chart/ -f values/image-to-pdf.yaml --namespace image-to-pdf-ns 
    helm install file-uploader ./helm-chart/ -f values/file-uploader.yaml --namespace file-uploader-ns  --set-file keyJson=./key.json
    ```
9. **Deploy  Istio Gateway and Virtual Service to route traffic to frontend**
    ```sh
    kubectl apply -f k8s-templates/istio/gateway.yaml
    ```

10. **Port Forward services to access them locally**:
    ```sh
    # Website Frontend 
    kubectl port-forward svc/istio-ingressgateway  -n istio-system 8080:80
    # Logs Dashboard 
    kubectl port-forward svc/kibana-kibana -n logging 5601:5601
    # Traces Dashboard
    kubectl port-forward svc/zipkin -n observability-ns 9411:9411
    # start kiali dashboard 
    istioctl dashboard kiali
    ``` 

11. **Access the services**:
    - Frontend: `http://localhost:8080`
    - Text-to-Speech: `http://localhost:8081`
    - Video-to-Audio: `http://localhost:8082`
    - Image-to-PDF: `http://localhost:8083`
    - File Uploader: `http://localhost:8084`
    - Kibana: `http://localhost:5601`
    - Zipkin: `http://localhost:9411`
    - Istio Kiali Dashboard:`http://localhost:20001`

## Usage

1. **Frontend**:
    - Open the frontend URL in your browser: **http://localhost:8080**
    - Use the provided interface to upload files and perform conversions.

2. **API Endpoints**:
    - Each service exposes gRPC endpoints for performing conversions. Refer to the respective service's proto files for detailed API documentation.

## Contact Information

For support or questions, or if you face any issues, please contact [gudikotieswarmani@gmail.com](mailto:gudikotieswarmani@gmail.com).

## Contributing

We welcome contributions to improve the project. Please fork the repository and submit pull requests.



