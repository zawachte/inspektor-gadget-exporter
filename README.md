# inspektor-gadget-exporter

inspektor-gadget-exporter is a prometheus exporter for inspektor-gadget. It launches gadgets and converts their outputs into prometheus/openmetrics style metrics.

## Usage

### Prerequisites 

* A kubernetes cluster.
* [inspektor-gadget](https://www.inspektor-gadget.io/) properly installed.

### Installation
The quickest way to get started with `inspektor-gadget-exporter` is to use the sample manifest included in this repo.

```bash
kubectl apply -f https://raw.githubusercontent.com/zawachte/inspektor-gadget-exporter/main/inspektor-gadget-exporter.yaml
```

### Validation

Start a port-forward to the inspektor-gadget-exporter service.

```bash
kubectl port-forward svc/inspektor-gadget-exporter 2112:2112
```

curl the metrics endpoint in another shell.
```bash
curl localhost:2112/metrics
```

## Build

```
make
```