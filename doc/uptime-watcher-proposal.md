# Uptime Watcher: 2-Day DevOps Mini Project Proposal

**Goal:** Build a small app and take it through a full DevOps flow: Git, Docker, CI/CD with Azure DevOps, Azure with Terraform, monitoring, and Kubernetes with Helm.

**Time:** 2 full days (about 8 to 9 hours each)

**Stack:** Go, Next.js, PostgreSQL, Docker, GitHub, Azure DevOps, Azure, Terraform, Helm

---

## 1. What We Are Building

A small app that checks if websites are up and shows their status on a dashboard.

- **Backend (Go):**
  - `POST /sites` to add a site
  - `GET /sites` to list sites with latest status
  - `GET /health` for health checks
  - A background goroutine that pings each site every minute
- **Frontend (Next.js):** one page showing each site as green (up) or red (down)
- **Database (PostgreSQL):** two tables, `sites` and `checks`

The app should take 2 hours at most. Keep it simple. The DevOps work is the real goal.

---

## 2. Architecture

```
Developer -> GitHub -> Azure Pipelines -> Azure Container Registry
                              |                      |
                              v                      v
                      Terraform (infra)     Azure Container Apps
                                                     |
                                    PostgreSQL + Key Vault
                                                     |
                                              Azure Monitor
                                          (logs, dashboard, alert)
```

- Code lives in **GitHub**
- **Azure Pipelines** runs CI/CD
- **ACR** stores Docker images
- **Terraform** creates all Azure resources
- **Key Vault** holds secrets
- **Azure Monitor** handles logs, dashboard, and alerts
- **Helm + minikube** is the local Kubernetes practice

---

## 3. Before You Start (30 minutes)

Do these before Day 1 so you don't get blocked.

- [ ] Install: Go, Node.js, Docker, Git, Azure CLI, Terraform, kubectl, minikube, Helm, Trivy
- [ ] Use your **new personal Azure free account**. It gives a $200 credit valid for 30 days, so do the project soon after sign-up
- [ ] Create an Azure DevOps organization and project with the same personal account
- [ ] **Parallel jobs:** new Azure DevOps organizations often have no free Microsoft-hosted agents, and the request form takes 2 to 3 business days. To avoid waiting, this plan uses a **self-hosted agent on an Azure VM** (see Section 4)
- [ ] Create an empty GitHub repo named `uptime-watcher`

---

## 4. Self-Hosted Agent on an Azure VM

The pipeline runs on your own VM, so you don't wait for Microsoft to approve free parallel jobs.

### Which VM size

| Size | RAM | Approx. price (East US, Linux) | Verdict |
|------|-----|--------------------------------|---------|
| B2ats v2 (free tier) | 1 GB | about $0.0094/hr (free for 12 months on a new account) | Too small: Docker and Next.js builds run out of memory |
| **B2als v2** | 4 GB | about $0.04/hr (estimate, confirm in the portal) | **Recommended: lowest suitable size** |
| B2as v2 | 8 GB | about $0.075/hr | Safe fallback if B2als v2 is not available or builds feel slow |

Do **not** pick B2pts v2. It is Arm-based, so your images would be built for Arm, but Azure Container Apps expects x64 (amd64) images.

Prices change, so check the exact price in the Azure portal when you create the VM.

### Setup (about 30 minutes, start of Day 1)

1. In the Azure portal, create an **Ubuntu 22.04 or 24.04** VM with size **B2als v2**
2. Use **SSH key** login, and allow port 22 only from your own IP
3. SSH in and install the tools:
   ```bash
   sudo apt update && sudo apt install -y docker.io git unzip
   sudo usermod -aG docker $USER
   ```
   Then install Trivy using its official apt repo instructions
4. Add a 2 GB swap file as a safety net:
   ```bash
   sudo fallocate -l 2G /swapfile
   sudo chmod 600 /swapfile
   sudo mkswap /swapfile && sudo swapon /swapfile
   echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
   ```
5. In Azure DevOps go to **Project Settings > Agent pools > Default > New agent**, download the Linux agent on the VM, then run:
   ```bash
   ./config.sh
   sudo ./svc.sh install && sudo ./svc.sh start
   ```
6. Create a **Personal Access Token (PAT)** with "Agent Pools (read, manage)" permission just for this setup. Do not save it anywhere afterwards.

### Rules for the VM

- Run minikube and Helm on your **laptop**, not on this VM
- Build images **one at a time**, not in parallel
- **Stop it as "Deallocated"** (not just shut down inside Linux) when you are not working, so compute billing stops
- **Delete the VM and its disk** when the project is finished
- Optional: if builds are still slow, switch the image build step to `az acr build`, which builds the image inside Azure instead of on the VM

---

## 5. Cost and Free Credit

You are using a new personal Azure free account, which gives:

- A **$200 credit valid for 30 days**, which can be used on services outside the free list, such as the B2als v2 VM
- Free monthly amounts of some services for 12 months

**Rough cost of this project**

- VM: about $2 for 48 hours (less if you deallocate it at night)
- PostgreSQL, Container Apps, ACR, Key Vault, and monitoring also use the credit, but two days of use should be well below $200

**How to stay safe**

- [ ] Check **Cost Management** in the portal at the start and end of each day
- [ ] Set a **budget alert** (this is already in the Day 2 plan)
- [ ] Run `terraform destroy` and delete the VM when you finish
- [ ] Do not leave anything running after the 30 days, because you would need to upgrade to pay-as-you-go to keep paid services

---

## 6. Repo Layout

```
uptime-watcher/
├── backend/             # Go API
├── frontend/            # Next.js app
├── infra/               # Terraform files
├── helm/                # Helm chart
├── azure-pipelines.yml  # CI/CD pipeline
├── docker-compose.yml
├── .env.example
├── .gitignore
└── README.md
```

---

## 7. Day 1: App, Docker, CI/CD

| Time | Task | Done when |
|------|------|-----------|
| 0:00-0:30 | Create the Azure VM (B2als v2), install tools, register the DevOps agent (Section 4) | Agent shows **Online** in Azure DevOps |
| 0:30-2:30 | Build the app (backend, frontend, DB) | Works locally |
| 2:30-3:00 | Git setup, branch rule on `main`, `.gitignore`, `.env` | `.env` is not in Git |
| 3:00-5:00 | Multi-stage Dockerfiles and `docker-compose.yml` | `docker compose up` runs everything, Go image under 30 MB |
| 5:00-5:45 | Lunch and buffer | |
| 5:45-8:45 | Azure DevOps pipeline | See checklist below |

### Pipeline setup steps

1. **Create work items** in Azure Boards (5 to 6, one per main task). Link commits with `AB#123` in the commit message.
2. **Create ACR** by hand (or with a tiny Terraform file). The pipeline needs it today. On Day 2 you bring it under full Terraform.
3. **Connect GitHub:** create a new pipeline, choose GitHub, and install the Azure Pipelines GitHub App on your repo.
4. **Create service connections** (Project Settings > Service connections):
   - One to your Azure subscription
   - One to ACR (name it `acr-connection`)
5. **Write `azure-pipelines.yml`** with these stages: Validate, Build and Scan, Push.
6. **Require the pipeline** as a status check on `main` in GitHub branch protection, so it runs on every pull request.

### Pipeline checklist

- [ ] Runs on every pull request
- [ ] Lint, then test, then build images
- [ ] Trivy scan fails the pipeline on HIGH or CRITICAL issues
- [ ] Merge to `main` pushes tagged images to ACR (tag: `$(Build.BuildId)`)

### Pipeline skeleton (backend and frontend)

```yaml
trigger:
  branches:
    include: [main]

pr:
  branches:
    include: [main]

variables:
  acr: youracrname            # change this
  acrConnection: acr-connection

pool:
  name: Default               # your self-hosted agent on the Azure VM

stages:
- stage: Validate
  jobs:
  - job: Backend
    steps:
    - script: cd backend && go vet ./... && go test ./...
      displayName: Backend lint and test
  - job: Frontend
    steps:
    - script: cd frontend && npm ci && npm run lint && npm test --if-present
      displayName: Frontend lint and test

- stage: BuildScan
  dependsOn: Validate
  jobs:
  - job: Images
    steps:
    - script: |
        docker build -t $(acr).azurecr.io/backend:$(Build.BuildId) ./backend
        docker build -t $(acr).azurecr.io/frontend:$(Build.BuildId) ./frontend
      displayName: Build images
    - script: |
        trivy image --exit-code 1 --severity HIGH,CRITICAL $(acr).azurecr.io/backend:$(Build.BuildId)
        trivy image --exit-code 1 --severity HIGH,CRITICAL $(acr).azurecr.io/frontend:$(Build.BuildId)
      displayName: Trivy scan
    - task: Docker@2
      displayName: Push backend
      condition: eq(variables['Build.SourceBranch'], 'refs/heads/main')
      inputs:
        command: push
        containerRegistry: $(acrConnection)
        repository: backend
        tags: $(Build.BuildId)
    - task: Docker@2
      displayName: Push frontend
      condition: eq(variables['Build.SourceBranch'], 'refs/heads/main')
      inputs:
        command: push
        containerRegistry: $(acrConnection)
        repository: frontend
        tags: $(Build.BuildId)
```

This is a starting point. Adjust names and paths for your project.

---

## 8. Day 2: Azure, Terraform, Monitoring, Helm

| Time | Task | Done when |
|------|------|-----------|
| 0:00-3:30 | Terraform: resource group, ACR, PostgreSQL, Key Vault, Container Apps | `terraform apply` builds everything from zero |
| 3:30-4:30 | Secrets, tags, budget alert, variable group | DB password comes from Key Vault, resources tagged, budget alert set |
| 4:30-5:30 | Lunch and buffer | |
| 5:30-7:00 | Monitoring | Logs in Azure Monitor, dashboard, email alert tested |
| 7:00-9:00 | Kubernetes and Helm | Chart deploys to minikube, one upgrade, one rollback |
| After | README, diagram, demo, `terraform destroy` | Everything cleaned up |

### Terraform

- Put everything in `infra/`: resource group, ACR, PostgreSQL Flexible Server, Key Vault, Container Apps environment, and the two Container Apps (backend, frontend)
- Use variables for names and region, and tag every resource (`project`, `owner`, `env`)
- Store Terraform state in an Azure Storage account (remote backend), not on your laptop
- Add a pipeline stage: `terraform plan`, then `terraform apply` behind a **manual approval** (use an Azure DevOps Environment with an approval check)

### Secrets

- Put the DB password and connection string in Key Vault
- Create a **variable group linked to Key Vault** in Azure DevOps so the pipeline reads secrets without storing them
- Container Apps read secrets from Key Vault using managed identity

### Deploy stage

- After the image is pushed, the pipeline updates the Container Apps with the new image tag
- Deployment happens from the pipeline, not from your laptop

### Monitoring

- Enable logs to a Log Analytics workspace
- Build a dashboard: CPU, memory, request count, error count
- Create an alert: email when `/health` fails or the app is down
- **Test the alert** by stopping the app

### Kubernetes and Helm (minikube)

- **Tip:** start minikube early (while Terraform is running) so you don't wait
- Create a chart in `helm/` for backend, frontend, and database
- Make `values-dev.yaml` (1 replica) and `values-prod.yaml` (3 replicas)
- Run `helm install`, then `helm upgrade` (change the image tag), then `helm rollback`
- Optional: add `helm lint` to the pipeline

---

## 9. Final Checklist

- [ ] Azure VM agent online and running the pipeline
- [ ] Multi-stage Dockerfiles, one-command local start
- [ ] Azure DevOps pipeline with tests and Trivy scan, linked to GitHub
- [ ] Images pushed to ACR on merge to `main`
- [ ] Branch protection with required pipeline check
- [ ] `terraform apply` and `terraform destroy` both work
- [ ] Terraform runs from the pipeline with manual approval
- [ ] No passwords in code, all from Key Vault
- [ ] Resources tagged and budget alert set
- [ ] `/health` endpoint, dashboard, down-alert tested
- [ ] Helm chart with `values-dev.yaml` and `values-prod.yaml`
- [ ] Azure Boards work items linked to commits
- [ ] README with architecture diagram and how to run it
- [ ] Short write-up: what broke, how you fixed it, what you would improve
- [ ] VM and all Azure resources deleted at the end, cost checked in Cost Management

---

## 10. What You Will Learn

- Docker multi-stage builds and Docker Compose
- Azure Pipelines, service connections, environments, and approvals
- Variable groups and Key Vault integration
- Terraform for Azure and remote state
- Container security scanning with Trivy
- Azure Monitor logs, dashboards, and alerts
- Kubernetes basics and Helm releases (install, upgrade, rollback)
- Setting up a self-hosted build agent on a Linux VM
- Cost control: tags, budgets, and `terraform destroy`

---

## 11. How to Stay on Schedule

- **Use AI for boilerplate.** Dockerfiles, Terraform blocks, and pipeline YAML can be generated. Read and understand them, don't just paste.
- **Start slow Azure things first.** The database and Container Apps take several minutes. Run `terraform apply`, then work on monitoring or Helm while you wait.
- **If you run behind, cut in this order:**
  1. Staging Helm values
  2. Dashboard polish
  3. Helm entirely
- **Never cut** the pipeline, Terraform, or the alert. Those are the core.
- **Break one thing on purpose** (wrong password, bad image) and fix it. Even 15 minutes of this teaches a lot.
- **Run `terraform destroy` and delete the VM at the end** so you don't burn Azure credits.
- **Test the pipeline on the VM in the first hour of Day 1.** If it is too slow, resize to B2as v2 (takes a few minutes, nothing is lost).

---

## 12. Final Demo (10 minutes)

1. Show the running app on Azure
2. Push a small code change in a pull request
3. Watch the pipeline run: test, build, scan
4. Merge, approve, and see the new version go live
5. Stop the app and show the alert email
6. Show the Helm rollback on minikube
7. Walk through the write-up: what broke and what you learned

---

## 13. Kubernetes Basics: Cluster, Node, Pod

Simple notes for the Kubernetes and Helm part of the project.

### What is a cluster?

A **cluster** is a group of computers (nodes) that work together as one system. You give it work, and it decides which computer runs it.

**Cluster = group of VMs (nodes) + a control plane that manages them**

*Example:* A restaurant kitchen. You place an order, and the head chef (control plane) decides which cook (node) makes each dish.

### How it is designed

```
              You (kubectl / helm)
                     |
                     v
     +-------------------------------+
     |  Control plane (the brain)    |
     |  - API server (takes orders)  |
     |  - Scheduler (picks a node)   |
     |  - Stores desired state       |
     +---------------+---------------+
                     |
          +----------+----------+
          v                     v
    +-----------+         +-----------+
    |  Node 1   |         |  Node 2   |
    | [Pod:API] |         | [Pod:API] |
    | [Pod:Web] |         | [Pod:DB]  |
    +-----------+         +-----------+
```

| Level | What it is | Example |
|-------|-----------|---------|
| Cluster | The whole system | Your minikube |
| Control plane | The brain that manages everything | Decides where pods run |
| Node | A VM that has joined the cluster (VM + kubelet + container runtime) | Node 1 |
| Deployment | Says how many pods you want | `replicas: 3` |
| Pod | One running copy of your app | `backend-abc12` |
| Container | Your app inside the pod | The backend image |

**Remember:** Cluster > Nodes (VMs) > Pods > Containers.

### Node

- A **node** is a VM that has joined a cluster. A plain VM becomes a node when Kubernetes software (kubelet and a container runtime) is installed and joined to the cluster.
- In a **real cluster** (such as AKS), one VM runs one node.
- In **practice clusters** (minikube, kind), one machine can pretend to be several nodes. Example: `minikube start --nodes 3` gives 3 nodes on your one laptop. This is good for learning, but gives no real protection, because if the laptop dies, all nodes die.
- The Azure DevOps agent VM in this project is **not** a node. It is a normal VM.

### Pod

- A **pod** is the smallest unit Kubernetes runs. Most of the time, one pod = one container (one running copy of your app).
- Each pod gets its own IP address.
- Pods are **temporary**. If a pod dies, Kubernetes creates a new one with a new name and IP. Example: `backend-abc12` crashes and `backend-ghi56` takes its place.
- You almost never create pods by hand. You create a **Deployment**, and it keeps the right number of pods running.
- Do not store data inside a pod. Use a database or a persistent volume.

### Many pods on one node

- One node can run many pods. The **scheduler** picks a node with enough free CPU and memory.
- One pod always runs on exactly one node. For copies on different nodes, create more replicas.
- Good practice: spread copies of the same app across different nodes. Example: if both backend pods are on Node 1 and Node 1 dies, the backend is down. If they are on different nodes, the app keeps running.

### Why we use a cluster

| Benefit | Example |
|---------|---------|
| Self-healing | A pod crashes at 3 AM and a new one starts in seconds |
| Scaling | Change "2 copies" to "5 copies" with one number |
| No-downtime updates | Rolling update replaces old pods with new ones slowly |
| Spreads the risk | If one node dies, its pods move to another node |
| Desired state | You write `replicas: 3` and the cluster keeps it that way |

### Where clusters appear in this project

| Part | Cluster? | Who manages it |
|------|----------|----------------|
| Azure Container Apps (live app) | Yes, hidden (Kubernetes behind the scenes) | Azure |
| Minikube (Helm practice) | Yes, local, 1 node by default | You, on your laptop |
| AKS (Azure Kubernetes Service) | **Not used** (costs more and takes longer to set up) | |
| Azure DevOps agent VM | No, just one normal VM | You |

**Optional bonus:** after the project works, deploy the same Helm chart to a 1-node AKS cluster for an hour, then delete it right away so it does not use your credit.

### Try it on minikube

```bash
minikube start --nodes 2          # 2 nodes on your laptop
kubectl get nodes                 # see the nodes
kubectl get pods -o wide          # NODE column shows where each pod runs
kubectl delete pod <pod-name>     # delete one pod
kubectl get pods                  # a new pod appears automatically (self-healing)
helm install uptime ./helm -f helm/values-dev.yaml
```
