# Edge Metrics Frontend

Edge device management dashboard built with React Router v7.

## Features

- 🚀 Server-side rendering
- ⚡️ Hot Module Replacement (HMR)
- 📦 Asset bundling and optimization
- 🔄 Real-time device status monitoring
- 🔒 TypeScript by default
- 🎉 TailwindCSS for styling
- 📊 Edge device configuration management
- 🔗 Backend API integration via Kubernetes Service DNS

## Getting Started

### Installation

Install the dependencies:

```bash
npm install
```

### Development

Start the development server with HMR:

```bash
npm run dev
```

Your application will be available at `http://localhost:5173`.

## Building for Production

Create a production build:

```bash
npm run build
```

## Deployment

### Prerequisites

- Docker
- Kubernetes cluster
- kubectl configured

### Method 1: Automated Deployment (Recommended)

#### Build and Deploy

```bash
# Basic deployment (local image)
./scripts/deploy.sh

# With version tag
./scripts/deploy.sh v1.0.0

# With Docker registry
REGISTRY=myregistry.com ./scripts/deploy.sh v1.0.0

# Full options
NAMESPACE=monitoring REGISTRY=myregistry.com ./scripts/deploy.sh v1.0.0
```

**Environment Variables:**
- `NAMESPACE`: Kubernetes namespace (default: monitoring)
- `REGISTRY`: Docker registry address

#### Check Deployment

```bash
# Check pods
kubectl get pods -n monitoring -l app.kubernetes.io/name=frontend

# View logs
kubectl logs -n monitoring -l app.kubernetes.io/name=frontend --tail=50 -f

# Check service
kubectl get svc -n monitoring edge-metrics-front
```

#### Undeploy

```bash
# Basic undeploy
./scripts/undeploy.sh

# Force delete without confirmation
FORCE=true ./scripts/undeploy.sh

# Delete with Docker image
DELETE_IMAGE=true ./scripts/undeploy.sh
```

### Method 2: Manual Deployment

#### 1. Build Docker Image

```bash
docker build -t edge-metrics-front:latest .
```

#### 2. Deploy to Kubernetes

```bash
# Create namespace (if not exists)
kubectl create namespace monitoring

# Apply manifests
kubectl apply -f manifests/deployment.yaml
kubectl apply -f manifests/service.yaml
```

#### 3. Access the Application

- **NodePort**: `http://<NodeIP>:30080`
- **Port Forward** (for local testing):
  ```bash
  kubectl port-forward -n monitoring svc/edge-metrics-front 3000:3000
  ```

### Environment Variables

#### Application Runtime
- `API_URL`: Backend API URL (default: `/api`)
  - Development: Uses vite proxy → `http://localhost:8081`
  - Production: Set to Kubernetes Service DNS → `http://edge-metrics-server:8081`
- `PORT`: Server port (default: 3000)

#### Deployment Scripts
- `NAMESPACE`: Kubernetes namespace (default: monitoring)
- `REGISTRY`: Docker registry address
- `DELETE_IMAGE`: Delete Docker image on undeploy (default: false)
- `FORCE`: Skip confirmation prompt (default: false)

## Architecture

```
edge-metrics-front/
├── app/                        # Application code
│   ├── routes/                 # React Router pages
│   ├── components/             # Reusable components
│   ├── lib/                    # Utilities
│   │   └── api.ts             # Backend API client
│   └── types/                  # TypeScript types
├── manifests/                  # Kubernetes manifests
│   ├── deployment.yaml        # Frontend Deployment
│   └── service.yaml           # Frontend Service (NodePort)
├── scripts/                    # Deployment automation
│   ├── build.sh               # Docker image build
│   ├── deploy.sh              # Automated deployment
│   └── undeploy.sh            # Automated cleanup
├── Dockerfile                  # Multi-stage build
└── vite.config.ts             # Vite configuration (dev proxy)
```

## Backend Integration

The frontend connects to the backend API at:
- **Development**: `/api` (proxied to `localhost:8081` via vite.config.ts)
- **Production**: `http://edge-metrics-server:8081` (Kubernetes Service DNS)

The API base URL is configured in [app/lib/api.ts](app/lib/api.ts#L14) using the `API_URL` environment variable.

## Styling

This application uses [Tailwind CSS](https://tailwindcss.com/) for styling.

---

Built with React Router v7
