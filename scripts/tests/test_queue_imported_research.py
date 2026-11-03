import contextlib
import importlib.util
import io
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('queue_imported', Path(__file__).parents[1] / 'queue-imported-research.py')
queue = importlib.util.module_from_spec(spec)
spec.loader.exec_module(queue)


class QueueImportTest(unittest.TestCase):
    def run_helper(self, tasks, expected, apply=True):
        calls = []
        def urlopen(req, timeout):
            calls.append((req.method, req.full_url))
            if req.method == 'GET':
                result = {'data': tasks, 'meta': {'totalPages': 1}}
            else:
                result = {'data': {'status': 'queued'}}
            return contextlib.closing(io.BytesIO(json.dumps(result).encode()))
        with tempfile.TemporaryDirectory() as directory:
            source = Path(directory) / 'import.json'
            source.write_text(json.dumps({'tasks': expected}))
            args = ['queue-imported-research.py', str(source)] + (['--apply'] if apply else [])
            with patch.dict(queue.os.environ, {'SCHOLARSHIP_OS_BASE_URL': 'https://example.test', 'SCHOLARSHIP_OS_AGENT_KEY': 'test-key'}), patch.object(queue.sys, 'argv', args), patch.object(queue.urllib.request, 'urlopen', urlopen), contextlib.redirect_stdout(io.StringIO()):
                try:
                    queue.main()
                    error = None
                except RuntimeError as failure:
                    error = failure
        return calls, error

    def test_apply_queues_only_matched_eligible_tasks(self):
        tasks = [dict(id=str(index), title=state, status=state, taskType='scholarship_research', profileId='master') for index, state in enumerate(['draft', 'ready', 'queued', 'running', 'completed'])]
        tasks.append(dict(id='unrelated', title='Other', status='draft', taskType='scholarship_research', profileId='master'))
        expected = [dict(title=t['title'], taskType=t['taskType'], profileId=t['profileId']) for t in tasks[:-1]]
        calls, error = self.run_helper(tasks, expected)
        self.assertIsNone(error)
        self.assertEqual([url for method, url in calls if method == 'POST'], ['https://example.test/api/v1/agent/research-tasks/0/queue', 'https://example.test/api/v1/agent/research-tasks/1/queue'])

    def test_preview_never_mutates(self):
        task = dict(id='1', title='Scholarship', status='draft', taskType='scholarship_research', profileId='master')
        calls, error = self.run_helper([task], [task], apply=False)
        self.assertIsNone(error)
        self.assertTrue(all(method == 'GET' for method, _ in calls))

    def test_ambiguous_or_missing_matches_block_all_mutations(self):
        task = dict(id='1', title='Scholarship', status='draft', taskType='scholarship_research', profileId='master')
        for tasks in [[], [task, dict(task, id='2')], [dict(task, profileId='other')]]:
            calls, error = self.run_helper(tasks, [task])
            self.assertIsNotNone(error)
            self.assertTrue(all(method == 'GET' for method, _ in calls))

    def test_instructions_disambiguate_imported_copy(self):
        task = dict(id='1', title='Scholarship', status='draft', taskType='scholarship_research', profileId='master', instructions='Current import')
        calls, error = self.run_helper([task, dict(task, id='2', instructions='Old research')], [task])
        self.assertIsNone(error)
        self.assertEqual(sum(method == 'POST' for method, _ in calls), 1)
        self.assertTrue(calls[-1][1].endswith('/1/queue'))


if __name__ == '__main__':
    unittest.main()
