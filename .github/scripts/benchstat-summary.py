#!/usr/bin/env python3
"""Render significant benchstat CSV changes and enforce regression thresholds."""

import argparse
import csv
import re
import sys
from dataclasses import dataclass
from typing import Iterable


CHANGE_RE = re.compile(r"^([+-]\d+(?:\.\d+)?)%$")


@dataclass(frozen=True)
class Change:
    name: str
    metric: str
    base: str
    head: str
    pct: float
    pval: str


def parse_benchstat_rows(rows: Iterable[list[str]]) -> list[Change]:
    changes: list[Change] = []
    columns: tuple[str, int, int, int, int] | None = None
    found_table = False
    for row in rows:
        if not row:
            continue
        if row[0] == "" and "vs base" in row:
            metric_indexes = [
                index
                for index, value in enumerate(row)
                if value and value not in {"CI", "vs base", "P"}
            ]
            if len(metric_indexes) < 2:
                raise ValueError(f"invalid benchstat metric header: {row}")
            columns = (
                row[metric_indexes[0]],
                metric_indexes[0],
                metric_indexes[1],
                row.index("vs base"),
                row.index("P"),
            )
            found_table = True
            continue
        if columns is None or row[0] == "geomean":
            continue

        metric, base_index, head_index, change_index, pval_index = columns
        required_length = max(base_index, head_index, change_index, pval_index) + 1
        row.extend([""] * (required_length - len(row)))
        if not row[base_index] or not row[head_index]:
            continue
        match = CHANGE_RE.match(row[change_index])
        if match is None:
            continue
        changes.append(
            Change(
                name=row[0],
                metric=metric,
                base=format_metric(row[base_index], metric),
                head=format_metric(row[head_index], metric),
                pct=float(match.group(1)),
                pval=row[pval_index].removeprefix("p=").split()[0],
            )
        )
    if not found_table:
        raise ValueError("input contains no benchstat CSV metric tables")
    return changes


def parse_benchstat(path: str) -> list[Change]:
    with open(path, newline="", encoding="utf-8") as source:
        return parse_benchstat_rows(csv.reader(source))


def format_metric(raw: str, metric: str) -> str:
    value = float(raw)
    if metric == "sec/op":
        for scale, suffix in ((1, "s/op"), (1e3, "ms/op"), (1e6, "us/op"), (1e9, "ns/op")):
            converted = value * scale
            if converted >= 1:
                return f"{converted:.3g} {suffix}"
    if metric in {"B/op", "allocs/op"}:
        return f"{value:.3g} {metric}"
    return f"{value:.3g} {metric}"


def regressions_above_threshold(changes: list[Change], threshold: float) -> list[Change]:
    return sorted(
        [change for change in changes if change.pct > threshold],
        key=lambda change: -change.pct,
    )


def render_markdown(changes: list[Change], threshold: float) -> str:
    regressions = sorted([change for change in changes if change.pct > 0], key=lambda change: -change.pct)
    improvements = sorted([change for change in changes if change.pct < 0], key=lambda change: change.pct)
    if not regressions and not improvements:
        return "### No significant performance changes detected\n"

    lines: list[str] = []
    if regressions:
        above = regressions_above_threshold(changes, threshold)
        heading = f"### {len(regressions)} minor regression(s) (all within {threshold:g}% threshold)\n"
        if above:
            heading = f"### {len(regressions)} regression(s) detected (threshold: >{threshold:g}%)\n"
        lines.append(heading)
        lines.extend(render_table(regressions, threshold, emphasize_regressions=True))

    if improvements:
        lines.append("<details>")
        lines.append(f"<summary>{len(improvements)} improvement(s)</summary>\n")
        lines.extend(render_table(improvements, threshold, emphasize_regressions=False))
        lines.append("</details>")
        lines.append("")
    return "\n".join(lines)


def render_table(changes: list[Change], threshold: float, emphasize_regressions: bool) -> list[str]:
    lines = [
        "| Benchmark | Metric | Base | Head | Change | p-value |",
        "|-----------|--------|------|------|--------|---------|",
    ]
    for change in changes:
        percentage = f"{change.pct:+.2f}%"
        if emphasize_regressions and change.pct > threshold:
            percentage = f"**{percentage}**"
        lines.append(
            f"| `{change.name}` | {change.metric} | {change.base} | {change.head} | {percentage} | {change.pval} |"
        )
    lines.append("")
    return lines


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("input", help="Path to benchstat CSV output")
    parser.add_argument("--threshold", type=float, default=5, help="Regression threshold percentage")
    parser.add_argument("--no-fail", action="store_true", help="Render regressions without returning a failure")
    args = parser.parse_args()

    try:
        changes = parse_benchstat(args.input)
    except (OSError, ValueError) as error:
        parser.error(str(error))
    print(render_markdown(changes, args.threshold))

    regressions = regressions_above_threshold(changes, args.threshold)
    if regressions and not args.no_fail:
        print(f"\nFailed: {len(regressions)} metric(s) regressed by more than {args.threshold:g}%:")
        for change in regressions:
            print(f"  {change.name} {change.metric}: {change.base} -> {change.head} ({change.pct:+.2f}%)")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
