#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEST="$SCRIPT_DIR/testdata/test-repo"

if [ -d "$DEST" ]; then
  echo "Removing existing $DEST"
  rm -rf "$DEST"
fi

mkdir -p "$DEST"
cd "$DEST"
git init
git config user.name "Test User"
git config user.email "test@example.com"

# ============================================================
# M1: Initial commit - README.md
# ============================================================
cat > README.md <<'PYEOF'
# Task Tracker

A simple CLI task tracker written in Python.

## Usage

```
python cli.py add "Buy groceries"
python cli.py list
python cli.py done 1
```

## Storage

Tasks are stored in `tasks.json` in the current directory.
PYEOF

git add README.md
git commit -m "Initial commit"

# ============================================================
# M2: Add task tracker with basic CRUD
# ============================================================
cat > task.py <<'PYEOF'
"""Task model for the task tracker."""

from dataclasses import dataclass, field
from datetime import datetime


@dataclass
class Task:
    """Represents a single task."""
    id: int
    title: str
    done: bool = False
    created_at: str = field(default_factory=lambda: datetime.now().isoformat())

    def __str__(self):
        status = "x" if self.done else " "
        return f"[{status}] {self.id}: {self.title}"

    def to_dict(self):
        return {
            "id": self.id,
            "title": self.title,
            "done": self.done,
            "created_at": self.created_at,
        }

    @classmethod
    def from_dict(cls, data):
        return cls(
            id=data["id"],
            title=data["title"],
            done=data.get("done", False),
            created_at=data.get("created_at", ""),
        )
PYEOF

cat > storage.py <<'PYEOF'
"""JSON file storage for tasks."""

import json
import os

from task import Task

DEFAULT_FILE = "tasks.json"


def load_tasks(filepath=DEFAULT_FILE):
    """Load tasks from a JSON file."""
    if not os.path.exists(filepath):
        return []
    with open(filepath, "r") as f:
        data = json.load(f)
    return [Task.from_dict(item) for item in data]


def save_tasks(tasks, filepath=DEFAULT_FILE):
    """Save tasks to a JSON file."""
    data = [t.to_dict() for t in tasks]
    with open(filepath, "w") as f:
        json.dump(data, f, indent=2)


def next_id(tasks):
    """Return the next available task ID."""
    if not tasks:
        return 1
    return max(t.id for t in tasks) + 1
PYEOF

cat > cli.py <<'PYEOF'
"""CLI interface for the task tracker."""

import argparse
import sys

from task import Task
from storage import load_tasks, save_tasks, next_id


def cmd_add(args):
    """Add a new task."""
    tasks = load_tasks()
    task = Task(id=next_id(tasks), title=args.title)
    tasks.append(task)
    save_tasks(tasks)
    print(f"Added: {task}")


def cmd_list(args):
    """List all tasks."""
    tasks = load_tasks()
    if not tasks:
        print("No tasks found.")
        return
    print(f"{'ID':<4} {'Status':<8} {'Title'}")
    print("-" * 40)
    for task in tasks:
        status = "done" if task.done else "pending"
        print(f"{task.id:<4} {status:<8} {task.title}")


def cmd_done(args):
    """Mark a task as done."""
    tasks = load_tasks()
    for task in tasks:
        if task.id == args.id:
            task.done = True
            save_tasks(tasks)
            print(f"Completed: {task}")
            return
    print(f"Task {args.id} not found.", file=sys.stderr)
    sys.exit(1)


def main():
    parser = argparse.ArgumentParser(description="Simple task tracker")
    sub = parser.add_subparsers(dest="command")

    add_p = sub.add_parser("add", help="Add a new task")
    add_p.add_argument("title", help="Task title")

    sub.add_parser("list", help="List all tasks")

    done_p = sub.add_parser("done", help="Mark a task as done")
    done_p.add_argument("id", type=int, help="Task ID to mark as done")

    args = parser.parse_args()
    if args.command is None:
        parser.print_help()
        sys.exit(1)

    commands = {"add": cmd_add, "list": cmd_list, "done": cmd_done}
    commands[args.command](args)


if __name__ == "__main__":
    main()
PYEOF

git add task.py storage.py cli.py
git commit -m "Add task tracker with basic CRUD"

# ============================================================
# Create feature branch
# ============================================================
git checkout -b feature/add-priority

# ============================================================
# F1: Add priority field to Task model
# Review points:
#   - bare except: catches SystemExit/KeyboardInterrupt
#   - __str__ does not display the new priority field
# ============================================================
cat > task.py <<'PYEOF'
"""Task model for the task tracker."""

from dataclasses import dataclass, field
from datetime import datetime

VALID_PRIORITIES = ["low", "medium", "high"]


@dataclass
class Task:
    """Represents a single task."""
    id: int
    title: str
    done: bool = False
    priority: str = "medium"
    created_at: str = field(default_factory=lambda: datetime.now().isoformat())

    def __post_init__(self):
        try:
            self.priority = self.priority.lower()
        except:
            self.priority = "medium"

    def __str__(self):
        status = "x" if self.done else " "
        return f"[{status}] {self.id}: {self.title}"

    def to_dict(self):
        return {
            "id": self.id,
            "title": self.title,
            "done": self.done,
            "priority": self.priority,
            "created_at": self.created_at,
        }

    @classmethod
    def from_dict(cls, data):
        return cls(
            id=data["id"],
            title=data["title"],
            done=data.get("done", False),
            priority=data.get("priority", "medium"),
            created_at=data.get("created_at", ""),
        )
PYEOF

git add task.py
git commit -m "Add priority field to Task model"

# ============================================================
# F2: Support priority flag in CLI add command
# Review points:
#   - argparse --priority lacks choices=["low","medium","high"]
#   - help text does not list valid priority values
# ============================================================
cat > cli.py <<'PYEOF'
"""CLI interface for the task tracker."""

import argparse
import sys

from task import Task
from storage import load_tasks, save_tasks, next_id


def cmd_add(args):
    """Add a new task."""
    tasks = load_tasks()
    task = Task(id=next_id(tasks), title=args.title, priority=args.priority)
    tasks.append(task)
    save_tasks(tasks)
    print(f"Added: {task}")


def cmd_list(args):
    """List all tasks."""
    tasks = load_tasks()
    if not tasks:
        print("No tasks found.")
        return
    print(f"{'ID':<4} {'Status':<8} {'Title'}")
    print("-" * 40)
    for task in tasks:
        status = "done" if task.done else "pending"
        print(f"{task.id:<4} {status:<8} {task.title}")


def cmd_done(args):
    """Mark a task as done."""
    tasks = load_tasks()
    for task in tasks:
        if task.id == args.id:
            task.done = True
            save_tasks(tasks)
            print(f"Completed: {task}")
            return
    print(f"Task {args.id} not found.", file=sys.stderr)
    sys.exit(1)


def main():
    parser = argparse.ArgumentParser(description="Simple task tracker")
    sub = parser.add_subparsers(dest="command")

    add_p = sub.add_parser("add", help="Add a new task")
    add_p.add_argument("title", help="Task title")
    add_p.add_argument("--priority", "-p", default="medium", help="Task priority")

    sub.add_parser("list", help="List all tasks")

    done_p = sub.add_parser("done", help="Mark a task as done")
    done_p.add_argument("id", type=int, help="Task ID to mark as done")

    args = parser.parse_args()
    if args.command is None:
        parser.print_help()
        sys.exit(1)

    commands = {"add": cmd_add, "list": cmd_list, "done": cmd_done}
    commands[args.command](args)


if __name__ == "__main__":
    main()
PYEOF

git add cli.py
git commit -m "Support priority flag in CLI add command"

# ============================================================
# F3: Sort tasks by priority in list output
# Review points:
#   - PRIORITY_ORDER dict duplicates ordering logic vs VALID_PRIORITIES in task.py
#   - Column header 'Priority' (8 chars) uses <6 format, misaligned with data
# ============================================================
cat > storage.py <<'PYEOF'
"""JSON file storage for tasks."""

import json
import os

from task import Task

DEFAULT_FILE = "tasks.json"

PRIORITY_ORDER = {"high": 0, "medium": 1, "low": 2}


def load_tasks(filepath=DEFAULT_FILE):
    """Load tasks from a JSON file."""
    if not os.path.exists(filepath):
        return []
    with open(filepath, "r") as f:
        data = json.load(f)
    return [Task.from_dict(item) for item in data]


def save_tasks(tasks, filepath=DEFAULT_FILE):
    """Save tasks to a JSON file."""
    data = [t.to_dict() for t in tasks]
    with open(filepath, "w") as f:
        json.dump(data, f, indent=2)


def next_id(tasks):
    """Return the next available task ID."""
    if not tasks:
        return 1
    return max(t.id for t in tasks) + 1


def sort_by_priority(tasks):
    """Sort tasks by priority (high first)."""
    return sorted(tasks, key=lambda t: PRIORITY_ORDER.get(t.priority, 99))
PYEOF

cat > cli.py <<'PYEOF'
"""CLI interface for the task tracker."""

import argparse
import sys

from task import Task
from storage import load_tasks, save_tasks, next_id, sort_by_priority


def cmd_add(args):
    """Add a new task."""
    tasks = load_tasks()
    task = Task(id=next_id(tasks), title=args.title, priority=args.priority)
    tasks.append(task)
    save_tasks(tasks)
    print(f"Added: {task}")


def cmd_list(args):
    """List all tasks."""
    tasks = load_tasks()
    if not tasks:
        print("No tasks found.")
        return
    tasks = sort_by_priority(tasks)
    print(f"{'ID':<4} {'Status':<8} {'Priority':<6} {'Title'}")
    print("-" * 50)
    for task in tasks:
        status = "done" if task.done else "pending"
        print(f"{task.id:<4} {status:<8} {task.priority:<6} {task.title}")


def cmd_done(args):
    """Mark a task as done."""
    tasks = load_tasks()
    for task in tasks:
        if task.id == args.id:
            task.done = True
            save_tasks(tasks)
            print(f"Completed: {task}")
            return
    print(f"Task {args.id} not found.", file=sys.stderr)
    sys.exit(1)


def main():
    parser = argparse.ArgumentParser(description="Simple task tracker")
    sub = parser.add_subparsers(dest="command")

    add_p = sub.add_parser("add", help="Add a new task")
    add_p.add_argument("title", help="Task title")
    add_p.add_argument("--priority", "-p", default="medium", help="Task priority")

    sub.add_parser("list", help="List all tasks")

    done_p = sub.add_parser("done", help="Mark a task as done")
    done_p.add_argument("id", type=int, help="Task ID to mark as done")

    args = parser.parse_args()
    if args.command is None:
        parser.print_help()
        sys.exit(1)

    commands = {"add": cmd_add, "list": cmd_list, "done": cmd_done}
    commands[args.command](args)


if __name__ == "__main__":
    main()
PYEOF

git add storage.py cli.py
git commit -m "Sort tasks by priority in list output"

# ============================================================
# F4: Add priority filter to list command
# Review points:
#   - filter_by_priority uses case-sensitive == comparison (bug)
#   - filter_by_priority is a standalone function, should be a classmethod on Task
# ============================================================
cat > task.py <<'PYEOF'
"""Task model for the task tracker."""

from dataclasses import dataclass, field
from datetime import datetime

VALID_PRIORITIES = ["low", "medium", "high"]


@dataclass
class Task:
    """Represents a single task."""
    id: int
    title: str
    done: bool = False
    priority: str = "medium"
    created_at: str = field(default_factory=lambda: datetime.now().isoformat())

    def __post_init__(self):
        try:
            self.priority = self.priority.lower()
        except:
            self.priority = "medium"

    def __str__(self):
        status = "x" if self.done else " "
        return f"[{status}] {self.id}: {self.title}"

    def to_dict(self):
        return {
            "id": self.id,
            "title": self.title,
            "done": self.done,
            "priority": self.priority,
            "created_at": self.created_at,
        }

    @classmethod
    def from_dict(cls, data):
        return cls(
            id=data["id"],
            title=data["title"],
            done=data.get("done", False),
            priority=data.get("priority", "medium"),
            created_at=data.get("created_at", ""),
        )


def filter_by_priority(tasks, priority):
    """Filter tasks by priority level."""
    return [t for t in tasks if t.priority == priority]
PYEOF

cat > cli.py <<'PYEOF'
"""CLI interface for the task tracker."""

import argparse
import sys

from task import Task, filter_by_priority
from storage import load_tasks, save_tasks, next_id, sort_by_priority


def cmd_add(args):
    """Add a new task."""
    tasks = load_tasks()
    task = Task(id=next_id(tasks), title=args.title, priority=args.priority)
    tasks.append(task)
    save_tasks(tasks)
    print(f"Added: {task}")


def cmd_list(args):
    """List all tasks."""
    tasks = load_tasks()
    if not tasks:
        print("No tasks found.")
        return
    if args.priority:
        tasks = filter_by_priority(tasks, args.priority)
    tasks = sort_by_priority(tasks)
    print(f"{'ID':<4} {'Status':<8} {'Priority':<6} {'Title'}")
    print("-" * 50)
    for task in tasks:
        status = "done" if task.done else "pending"
        print(f"{task.id:<4} {status:<8} {task.priority:<6} {task.title}")


def cmd_done(args):
    """Mark a task as done."""
    tasks = load_tasks()
    for task in tasks:
        if task.id == args.id:
            task.done = True
            save_tasks(tasks)
            print(f"Completed: {task}")
            return
    print(f"Task {args.id} not found.", file=sys.stderr)
    sys.exit(1)


def main():
    parser = argparse.ArgumentParser(description="Simple task tracker")
    sub = parser.add_subparsers(dest="command")

    add_p = sub.add_parser("add", help="Add a new task")
    add_p.add_argument("title", help="Task title")
    add_p.add_argument("--priority", "-p", default="medium", help="Task priority")

    list_p = sub.add_parser("list", help="List all tasks")
    list_p.add_argument("--priority", "-p", help="Filter by priority")

    done_p = sub.add_parser("done", help="Mark a task as done")
    done_p.add_argument("id", type=int, help="Task ID to mark as done")

    args = parser.parse_args()
    if args.command is None:
        parser.print_help()
        sys.exit(1)

    commands = {"add": cmd_add, "list": cmd_list, "done": cmd_done}
    commands[args.command](args)


if __name__ == "__main__":
    main()
PYEOF

git add task.py cli.py
git commit -m "Add priority filter to list command"

echo ""
echo "=== Test repo generated at $DEST ==="
echo "Branch: $(git branch --show-current)"
echo ""
echo "Main branch commits:"
git log --oneline main
echo ""
echo "Feature branch commits (main..HEAD):"
git log --oneline main..HEAD
