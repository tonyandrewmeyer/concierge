#!/bin/bash
# debug-metallb.sh - Minimal reproducer / diagnostic script for metallb enable timeout.
#
# This script:
#   1. Collects pre-enable cluster state
#   2. Runs `microk8s enable metallb:...` in the background
#   3. Periodically collects cluster state while waiting
#   4. Reports final state on success, timeout, or failure
#
# Usage: sudo bash debug-metallb.sh [timeout_seconds]
#   Default timeout: 600 (10 minutes)

set -uo pipefail

TIMEOUT="${1:-600}"
METALLB_RANGE="10.64.140.43-10.64.140.49"
LOG_DIR="/tmp/metallb-debug-$(date +%Y%m%d-%H%M%S)"
POLL_INTERVAL=15

mkdir -p "$LOG_DIR"
echo "=== MetalLB debug session started at $(date -u) ==="
echo "=== Logs: $LOG_DIR ==="
echo "=== Timeout: ${TIMEOUT}s ==="

collect_state() {
    local label="$1"
    local dir="${LOG_DIR}/${label}"
    mkdir -p "$dir"

    echo "--- Collecting state: ${label} ($(date -u)) ---"

    # Node info
    microk8s kubectl get nodes -o wide > "$dir/nodes.txt" 2>&1 || true

    # All pods across all namespaces
    microk8s kubectl get pods -A -o wide > "$dir/pods-all.txt" 2>&1 || true

    # MetalLB-specific namespace pods (may not exist yet)
    microk8s kubectl get pods -n metallb-system -o wide > "$dir/pods-metallb.txt" 2>&1 || true
    microk8s kubectl describe pods -n metallb-system > "$dir/describe-pods-metallb.txt" 2>&1 || true

    # Events in metallb-system namespace
    microk8s kubectl get events -n metallb-system --sort-by='.lastTimestamp' > "$dir/events-metallb.txt" 2>&1 || true

    # Events cluster-wide (recent)
    microk8s kubectl get events -A --sort-by='.lastTimestamp' > "$dir/events-all.txt" 2>&1 || true

    # Deployments and daemonsets in metallb-system
    microk8s kubectl get deploy,daemonset,replicaset -n metallb-system -o wide > "$dir/workloads-metallb.txt" 2>&1 || true

    # Check for any pending container images
    microk8s kubectl get pods -n metallb-system -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{range .status.containerStatuses[*]}{.state}{"\n"}{end}{end}' > "$dir/container-states.txt" 2>&1 || true

    # Helm releases (metallb is deployed via helm in recent microk8s)
    microk8s helm3 ls -A > "$dir/helm-releases.txt" 2>&1 || true

    # Check image pull status
    microk8s ctr images list 2>/dev/null | grep -i metallb > "$dir/containerd-images.txt" 2>&1 || true

    # Pod logs if pods exist (current + previous for crash-looping containers)
    for pod in $(microk8s kubectl get pods -n metallb-system -o name 2>/dev/null); do
        local pod_name="${pod#pod/}"
        microk8s kubectl logs "$pod" -n metallb-system --all-containers > "$dir/logs-${pod_name}.txt" 2>&1 || true
        microk8s kubectl logs "$pod" -n metallb-system --all-containers --previous > "$dir/logs-${pod_name}-previous.txt" 2>&1 || true
    done

    # MicroK8s addon status
    microk8s status --format yaml > "$dir/microk8s-status.yaml" 2>&1 || true

    # System resource info
    free -h > "$dir/memory.txt" 2>&1 || true
    df -h > "$dir/disk.txt" 2>&1 || true
    nproc > "$dir/cpus.txt" 2>&1 || true

    # Network state
    ip addr > "$dir/ip-addr.txt" 2>&1 || true
    ss -tlnp > "$dir/listening-ports.txt" 2>&1 || true

    # DNS resolution check (unique pod name per collection to avoid conflicts)
    local dns_pod="dns-test-$(date +%s)"
    timeout 30 microk8s kubectl run -n default --rm -i --restart=Never --image=busybox "$dns_pod" -- nslookup kubernetes.default.svc.cluster.local > "$dir/dns-test.txt" 2>&1 || true

    # Check snap microk8s processes
    ps auxf | grep -E "microk8s|helm|metallb" > "$dir/processes.txt" 2>&1 || true

    # Journal logs for microk8s (last 100 lines)
    journalctl -u snap.microk8s.daemon-kubelite --no-pager -n 100 > "$dir/journal-kubelite.txt" 2>&1 || true
    journalctl -u snap.microk8s.daemon-containerd --no-pager -n 100 > "$dir/journal-containerd.txt" 2>&1 || true

    echo "--- State collected: ${label} ---"
}

print_summary() {
    local label="$1"
    local dir="${LOG_DIR}/${label}"

    echo ""
    echo "=== Summary for ${label} ==="

    if [ -f "$dir/nodes.txt" ]; then
        echo "-- Nodes --"
        cat "$dir/nodes.txt"
    fi

    if [ -f "$dir/pods-all.txt" ]; then
        echo "-- All Pods --"
        cat "$dir/pods-all.txt"
    fi

    if [ -f "$dir/describe-pods-metallb.txt" ]; then
        echo "-- MetalLB Pod Descriptions --"
        cat "$dir/describe-pods-metallb.txt"
    fi

    if [ -f "$dir/events-metallb.txt" ]; then
        echo "-- MetalLB Events --"
        cat "$dir/events-metallb.txt"
    fi

    if [ -f "$dir/workloads-metallb.txt" ]; then
        echo "-- MetalLB Workloads --"
        cat "$dir/workloads-metallb.txt"
    fi

    if [ -f "$dir/helm-releases.txt" ]; then
        echo "-- Helm Releases --"
        cat "$dir/helm-releases.txt"
    fi

    if [ -f "$dir/containerd-images.txt" ]; then
        echo "-- MetalLB Container Images --"
        cat "$dir/containerd-images.txt"
    fi

    if [ -f "$dir/memory.txt" ]; then
        echo "-- Memory --"
        cat "$dir/memory.txt"
    fi

    # Print pod logs for metallb pods (current and previous)
    for logfile in "$dir"/logs-*.txt; do
        if [ -f "$logfile" ] && [ -s "$logfile" ]; then
            local log_name
            log_name="$(basename "$logfile" .txt | sed 's/^logs-//')"
            echo "-- Pod logs: ${log_name} --"
            cat "$logfile"
        fi
    done

    # Print logs for other crash-looping kube-system pods (coredns, calico, etc.)
    if [ -f "$dir/pods-all.txt" ]; then
        for crashing_pod in $(awk '/CrashLoopBackOff|Error|CreateContainerConfigError/ {print $1 ":" $2}' "$dir/pods-all.txt" 2>/dev/null); do
            local ns="${crashing_pod%%:*}"
            local pod="${crashing_pod#*:}"
            if [ "$ns" != "metallb-system" ] && [ -n "$pod" ]; then
                echo "-- Pod logs: ${ns}/${pod} --"
                microk8s kubectl logs -n "$ns" "$pod" --all-containers --tail=50 2>&1 || true
                echo "-- Previous pod logs: ${ns}/${pod} --"
                microk8s kubectl logs -n "$ns" "$pod" --all-containers --previous --tail=50 2>&1 || true
            fi
        done
    fi

    echo "=== End summary for ${label} ==="
}

# --- Pre-enable state ---
echo ""
echo "========================================"
echo "  PRE-ENABLE STATE"
echo "========================================"
collect_state "01-pre-enable"
print_summary "01-pre-enable"

# --- Run metallb enable in background ---
echo ""
echo "========================================"
echo "  ENABLING METALLB (timeout: ${TIMEOUT}s)"
echo "========================================"
echo "Running: microk8s enable metallb:${METALLB_RANGE}"
echo "Start time: $(date -u)"

microk8s enable "metallb:${METALLB_RANGE}" > "${LOG_DIR}/metallb-enable-stdout.txt" 2>&1 &
ENABLE_PID=$!
echo "PID: ${ENABLE_PID}"

# --- Poll while waiting ---
elapsed=0
poll_count=2
while kill -0 "$ENABLE_PID" 2>/dev/null; do
    if [ "$elapsed" -ge "$TIMEOUT" ]; then
        echo ""
        echo "========================================"
        echo "  TIMEOUT REACHED (${TIMEOUT}s)"
        echo "========================================"

        # Collect state at timeout
        collect_state "99-timeout"
        print_summary "99-timeout"

        echo ""
        echo "-- metallb enable stdout so far --"
        cat "${LOG_DIR}/metallb-enable-stdout.txt" || true

        echo ""
        echo "-- Process tree of metallb enable --"
        pstree -p "$ENABLE_PID" 2>/dev/null || true
        ps auxf | grep -E "(microk8s|metallb|helm|kubectl)" || true

        echo ""
        echo "-- Strace snapshot (2 seconds) --"
        timeout 2 strace -f -p "$ENABLE_PID" -e trace=network,write 2>&1 | head -100 || true

        # Kill the hung process
        echo ""
        echo "Killing metallb enable (PID: ${ENABLE_PID})..."
        kill "$ENABLE_PID" 2>/dev/null || true
        sleep 2
        kill -9 "$ENABLE_PID" 2>/dev/null || true

        echo ""
        echo "=== DEBUG SESSION COMPLETE (TIMEOUT) ==="
        echo "=== Full logs in: ${LOG_DIR} ==="
        exit 1
    fi

    sleep "$POLL_INTERVAL"
    elapsed=$((elapsed + POLL_INTERVAL))

    collect_state "$(printf '%02d' $poll_count)-poll-${elapsed}s"

    # Print a compact status line
    echo "[${elapsed}s] pods in metallb-system:"
    microk8s kubectl get pods -n metallb-system --no-headers 2>/dev/null || echo "  (none)"

    poll_count=$((poll_count + 1))
done

# --- Process completed ---
wait "$ENABLE_PID"
exit_code=$?
echo ""
echo "========================================"
echo "  METALLB ENABLE COMPLETED (exit code: ${exit_code})"
echo "========================================"
echo "End time: $(date -u)"
echo "Elapsed: ${elapsed}s"

echo ""
echo "-- metallb enable stdout --"
cat "${LOG_DIR}/metallb-enable-stdout.txt" || true

collect_state "99-post-enable"
print_summary "99-post-enable"

echo ""
echo "=== DEBUG SESSION COMPLETE (exit code: ${exit_code}) ==="
echo "=== Full logs in: ${LOG_DIR} ==="
exit "$exit_code"
