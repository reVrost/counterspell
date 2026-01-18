<script lang="ts">
	import FileCodeIcon from '@lucide/svelte/icons/file-code';

	interface Props {
		diff: string;
	}

	let { diff }: Props = $props();

	interface DiffFile {
		name: string;
		lines: DiffLine[];
	}

	interface DiffLine {
		type: 'add' | 'del' | 'context' | 'hunk';
		content: string;
		lineNum: string;
	}

	function parseDiff(diffText: string): DiffFile[] {
		const files: DiffFile[] = [];
		let currentFile: DiffFile | null = null;
		let lineNum = 0;

		for (const line of diffText.split('\n')) {
			if (line.startsWith('diff --git')) {
				if (currentFile) {
					files.push(currentFile);
				}
				const parts = line.split(' b/');
				const name = parts.length > 1 ? parts[parts.length - 1] : '';
				currentFile = { name, lines: [] };
				lineNum = 0;
			} else if (!currentFile) {
				continue;
			} else if (line.startsWith('@@')) {
				currentFile.lines.push({ type: 'hunk', content: line, lineNum: '' });
			} else if (line.startsWith('---') || line.startsWith('+++') || line.startsWith('index ')) {
				// Skip meta
			} else if (line.startsWith('+')) {
				lineNum++;
				currentFile.lines.push({ type: 'add', content: line, lineNum: String(lineNum) });
			} else if (line.startsWith('-')) {
				currentFile.lines.push({ type: 'del', content: line, lineNum: '' });
			} else if (line !== '') {
				lineNum++;
				currentFile.lines.push({ type: 'context', content: line, lineNum: String(lineNum) });
			}
		}

		if (currentFile) {
			files.push(currentFile);
		}

		return files;
	}

	const files = $derived(parseDiff(diff));
</script>

{#each files as file}
	<div class="diff-file-header">
		<FileCodeIcon class="w-3 h-3 mr-2 text-gray-500 inline" />
		{file.name}
	</div>
	<div class="diff-file-body">
		{#each file.lines as line}
			{#if line.type === 'hunk'}
				<div class="diff-hunk">{line.content}</div>
			{:else if line.type === 'add'}
				<div class="diff-line diff-add">
					<span class="diff-line-num">{line.lineNum}</span>
					<span class="diff-line-content">{line.content}</span>
				</div>
			{:else if line.type === 'del'}
				<div class="diff-line diff-del">
					<span class="diff-line-num"></span>
					<span class="diff-line-content">{line.content}</span>
				</div>
			{:else}
				<div class="diff-line diff-context">
					<span class="diff-line-num">{line.lineNum}</span>
					<span class="diff-line-content">{line.content}</span>
				</div>
			{/if}
		{/each}
	</div>
{/each}
