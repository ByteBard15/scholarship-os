#!/usr/bin/env python3
"""Match an existing import to workspace tasks; never create or start research."""
import argparse
import json
import os
from pathlib import Path
import re
import sys
import urllib.error
import urllib.request


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('import_file', type=Path)
    parser.add_argument('--credentials-file', type=Path, help='Local requests.http containing @base and @agent_token; never printed')
    parser.add_argument('--apply', action='store_true', help='Queue matched draft/ready tasks; otherwise preview only')
    parser.add_argument('--report', type=Path)
    args = parser.parse_args()
    base, key = os.environ.get('SCHOLARSHIP_OS_BASE_URL'), os.environ.get('SCHOLARSHIP_OS_AGENT_KEY')
    if args.credentials_file:
        config = args.credentials_file.read_text()
        def setting(name):
            match = re.search(r'^@' + name + r'\s*=\s*(.+)$', config, re.M)
            return match.group(1).strip() if match else None
        base, key = base or setting('base'), key or setting('agent_token')
    if not base or not key:
        parser.error('Set SCHOLARSHIP_OS_BASE_URL and SCHOLARSHIP_OS_AGENT_KEY, or provide --credentials-file')
    source = json.loads(args.import_file.read_text())
    expected = source.get('tasks', [])
    if not expected:
        parser.error('Import file has no tasks')

    def call(path, method='GET'):
        req = urllib.request.Request(base.rstrip('/') + '/api/v1' + path, method=method, headers={'Authorization': 'Bearer ' + key})
        try:
            with urllib.request.urlopen(req, timeout=60) as response:
                return json.load(response)
        except urllib.error.HTTPError as error:
            # Do not echo headers, request bodies or credential-bearing config.
            if error.code == 404:
                raise RuntimeError('Agent task listing/queue route unavailable; deploy the updated API first.') from None
            raise RuntimeError(f'API returned HTTP {error.code}; inspect task state and agent research scope.') from None

    # Read every page before mutation, so moving draft tasks cannot shift pages.
    tasks, page = [], 1
    while True:
        response = call(f'/agent/research-tasks/list?page={page}&pageSize=100&sortBy=createdAt')
        tasks.extend(response['data'])
        if page >= response['meta']['totalPages']:
            break
        page += 1

    matched, issues = [], []
    seen = set()
    for item in expected:
        candidates = [task for task in tasks if task['title'].strip() == item['title'].strip()
                      and task['taskType'] == item['taskType']
                      and task.get('profileId') == item.get('profileId')]
        if len(candidates) > 1:
            candidates = [task for task in candidates if task.get('instructions') == item.get('instructions')]
        if len(candidates) != 1:
            issues.append({'title': item['title'], 'problem': 'missing or ambiguous match', 'candidateIds': [task['id'] for task in candidates]})
            continue
        task = candidates[0]
        if task['id'] in seen:
            issues.append({'title': item['title'], 'problem': 'duplicate import entry'})
            continue
        seen.add(task['id'])
        matched.append({'id': task['id'], 'title': task['title'], 'before': task['status'], 'result': 'preview'})
    report = {'apply': args.apply, 'expected': len(expected), 'matched': len(matched), 'issues': issues, 'tasks': matched}
    def persist():
        if args.report:
            args.report.write_text(json.dumps(report, indent=2) + '\n')
    persist()
    if issues:
        print(json.dumps(report, indent=2))
        raise RuntimeError('Resolve missing/ambiguous matches before queueing. No queue requests were sent.')
    for task in matched:
        if task['before'] in ('draft', 'ready'):
            if args.apply:
                try:
                    result = call('/agent/research-tasks/' + task['id'] + '/queue', 'POST')['data']
                    task['result'] = result['status']
                except RuntimeError:
                    task['result'] = 'queue request failed; verify current state before retry'
                    persist()
                    raise
            else:
                task['result'] = 'would queue'
        else:
            task['result'] = 'already queued' if task['before'] == 'queued' else 'left unchanged: ' + task['before']
        persist()
        print(task['title'] + ': ' + task['result'], flush=True)
    print(f"Matched {len(matched)} of {len(expected)} imported tasks. " + ('Applied eligible queue changes.' if args.apply else 'Preview only.'))


if __name__ == '__main__':
    try:
        main()
    except (RuntimeError, OSError, ValueError, KeyError) as error:
        print('Stopped: ' + str(error), file=sys.stderr)
        sys.exit(1)
