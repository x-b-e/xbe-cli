#!/usr/bin/env python3
import argparse
import subprocess
import sys
import time
from pathlib import Path
from typing import List, Optional

from common import load_config, project_root


def format_duration(seconds: float) -> str:
    if seconds < 0:
        seconds = 0
    total = int(seconds)
    hours = total // 3600
    minutes = (total % 3600) // 60
    secs = total % 60
    return f"{hours:02d}:{minutes:02d}:{secs:02d}"


def log(message: str) -> None:
    timestamp = time.strftime("%H:%M:%S")
    print(f"[{timestamp}] {message}")


def main() -> None:
    parser = argparse.ArgumentParser(description="Run a swarm of resource-map workers")
    parser.add_argument("--config", default="config.yaml", help="Path to config.yaml")
    parser.add_argument(
        "--agent",
        default="codex",
        choices=["codex", "claude"],
        help="Agent runner (codex or claude)",
    )
    parser.add_argument(
        "--agent-model",
        default="gpt-5.2-codex",
        help="Model name for agent runner",
    )
    parser.add_argument(
        "--reasoning",
        default="medium",
        choices=["low", "medium", "high", "minimal"],
        help="Reasoning effort for codex models (low/medium/high/minimal)",
    )
    parser.add_argument(
        "--concurrency",
        type=int,
        default=4,
        help="Number of concurrent workers",
    )
    parser.add_argument(
        "--max-batches",
        type=int,
        help="Maximum number of batches to process before exiting",
    )
    args = parser.parse_args()

    worker_path = Path(__file__).with_name("resource_map_worker.py")
    config = load_config(args.config)
    root_out = project_root(config)
    pending_dir = root_out / "resource_map" / "queue" / "pending"

    remaining: Optional[int] = args.max_batches
    exit_code = 0
    total_processed = 0
    start_time = time.time()

    if args.max_batches is None:
        pending = sorted(pending_dir.glob("batch_*.json"))
        if not pending:
            log("No pending batches remaining")
            sys.exit(exit_code)
        log(f"Launching {args.concurrency} loop worker(s); pending={len(pending)}")
        processes: List[Optional[subprocess.Popen]] = []

        def spawn_worker() -> subprocess.Popen:
            cmd = [
                sys.executable,
                str(worker_path),
                "--config",
                args.config,
                "--agent",
                args.agent,
                "--loop",
            ]
            if args.agent_model:
                cmd.extend(["--agent-model", args.agent_model])
            if args.reasoning:
                cmd.extend(["--reasoning", args.reasoning])
            return subprocess.Popen(cmd)

        for _ in range(max(args.concurrency, 1)):
            processes.append(spawn_worker())

        while True:
            pending = sorted(pending_dir.glob("batch_*.json"))
            any_running = False
            for idx, process in enumerate(processes):
                if process is None:
                    continue
                result = process.poll()
                if result is None:
                    any_running = True
                    continue
                if result != 0:
                    exit_code = result
                if pending:
                    log("Worker exited; respawning to maintain concurrency")
                    processes[idx] = spawn_worker()
                    any_running = True
                else:
                    processes[idx] = None
            if not pending and not any_running:
                break
            time.sleep(1)
        sys.exit(exit_code)

    while True:
        pending = sorted(pending_dir.glob("batch_*.json"))
        if not pending:
            log("No pending batches remaining")
            break
        if remaining is not None and remaining <= 0:
            log("Reached max-batches limit")
            break

        batch_slots = max(args.concurrency, 1)
        if remaining is not None:
            batch_slots = min(batch_slots, remaining)
        batch_slots = min(batch_slots, len(pending))
        if batch_slots <= 0:
            break

        prev_pending = len(pending)
        log(f"Launching {batch_slots} worker(s); pending={prev_pending}")
        processes: List[subprocess.Popen] = []
        for _ in range(batch_slots):
            cmd = [
                sys.executable,
                str(worker_path),
                "--config",
                args.config,
                "--agent",
                args.agent,
            ]
            if args.agent_model:
                cmd.extend(["--agent-model", args.agent_model])
            if args.reasoning:
                cmd.extend(["--reasoning", args.reasoning])
            processes.append(subprocess.Popen(cmd))

        cycle_start = time.time()
        for process in processes:
            result = process.wait()
            if result != 0:
                exit_code = result
        cycle_elapsed = time.time() - cycle_start

        if remaining is not None:
            remaining -= batch_slots

        current_pending = len(sorted(pending_dir.glob("batch_*.json")))
        processed = max(prev_pending - current_pending, 0)
        total_processed += processed
        elapsed = time.time() - start_time
        if total_processed > 0:
            avg = elapsed / total_processed
            remaining_batches = remaining if remaining is not None else current_pending
            eta = format_duration(avg * remaining_batches)
            rate = total_processed / elapsed if elapsed > 0 else 0
            log(
                f"Completed {processed} batch(es) in {format_duration(cycle_elapsed)}; "
                f"total={total_processed}, pending={current_pending}, "
                f"rate={rate:.2f}/s, ETA={eta}"
            )
        else:
            log(
                f"Completed {processed} batch(es) in {format_duration(cycle_elapsed)}; "
                f"pending={current_pending}"
            )

    sys.exit(exit_code)


if __name__ == "__main__":
    main()
