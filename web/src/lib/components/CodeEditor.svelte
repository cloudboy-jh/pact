<script lang="ts">
	import { onMount, onDestroy, createEventDispatcher } from 'svelte';
	import { EditorState } from '@codemirror/state';
	import { EditorView, keymap, highlightActiveLine, highlightActiveLineGutter, lineNumbers, highlightSpecialChars } from '@codemirror/view';
	import { json } from '@codemirror/lang-json';
	import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
	import { searchKeymap, highlightSelectionMatches } from '@codemirror/search';
	import { autocompletion, completionKeymap } from '@codemirror/autocomplete';
	import { bracketMatching, foldGutter, foldKeymap } from '@codemirror/language';
	
	export let content: string = '';
	export let language: 'json' | 'text' = 'json';
	export let readonly: boolean = false;
	export let highlightLines: { from: number; to: number } | null = null;

	const dispatch = createEventDispatcher<{ change: string }>();

	let editorContainer: HTMLDivElement;
	let view: EditorView | null = null;

	// Custom dark theme matching our UI
	const pactTheme = EditorView.theme({
		'&': {
			backgroundColor: 'transparent',
			color: '#e4e4e7', // zinc-200
			height: '100%'
		},
		'.cm-content': {
			caretColor: '#34d399', // emerald-400
			fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
			fontSize: '13px',
			lineHeight: '1.6'
		},
		'.cm-cursor': {
			borderLeftColor: '#34d399' // emerald-400
		},
		'&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection': {
			backgroundColor: '#3f3f46' // zinc-700
		},
		'.cm-activeLine': {
			backgroundColor: 'rgba(63, 63, 70, 0.3)' // zinc-700/30
		},
		'.cm-activeLineGutter': {
			backgroundColor: 'rgba(63, 63, 70, 0.3)' // zinc-700/30
		},
		'.cm-gutters': {
			backgroundColor: 'transparent',
			color: '#52525b', // zinc-600
			border: 'none',
			paddingRight: '8px'
		},
		'.cm-lineNumbers .cm-gutterElement': {
			padding: '0 8px 0 16px',
			minWidth: '40px'
		},
		'.cm-foldGutter .cm-gutterElement': {
			padding: '0 4px'
		},
		// JSON syntax highlighting
		'.cm-string': {
			color: '#34d399' // emerald-400
		},
		'.cm-number': {
			color: '#fbbf24' // amber-400
		},
		'.cm-bool': {
			color: '#60a5fa' // blue-400
		},
		'.cm-null': {
			color: '#71717a' // zinc-500
		},
		'.cm-propertyName': {
			color: '#f4f4f5' // zinc-100
		},
		'.cm-punctuation': {
			color: '#a1a1aa' // zinc-400
		},
		// Highlighted line (for section highlighting)
		'.cm-highlighted-line': {
			backgroundColor: 'rgba(52, 211, 153, 0.15)', // emerald-400/15
			borderLeft: '2px solid #34d399'
		}
	}, { dark: true });

	// Line highlighting decoration
	import { Decoration, type DecorationSet } from '@codemirror/view';
	import { StateField, StateEffect } from '@codemirror/state';

	const highlightEffect = StateEffect.define<{ from: number; to: number } | null>();

	const highlightField = StateField.define<DecorationSet>({
		create() {
			return Decoration.none;
		},
		update(decorations, tr) {
			for (const effect of tr.effects) {
				if (effect.is(highlightEffect)) {
					if (effect.value === null) {
						return Decoration.none;
					}
					const { from, to } = effect.value;
					const decorationList = [];
					for (let line = from; line <= to; line++) {
						try {
							const lineInfo = tr.state.doc.line(line);
							decorationList.push(
								Decoration.line({ class: 'cm-highlighted-line' }).range(lineInfo.from)
							);
						} catch {
							// Line doesn't exist
						}
					}
					return Decoration.set(decorationList);
				}
			}
			return decorations;
		},
		provide: f => EditorView.decorations.from(f)
	});

	function getExtensions() {
		const extensions = [
			pactTheme,
			lineNumbers(),
			highlightActiveLineGutter(),
			highlightSpecialChars(),
			history(),
			foldGutter(),
			bracketMatching(),
			highlightActiveLine(),
			highlightSelectionMatches(),
			highlightField,
			keymap.of([
				...defaultKeymap,
				...historyKeymap,
				...searchKeymap,
				...foldKeymap,
				...completionKeymap
			])
		];

		if (language === 'json') {
			extensions.push(json());
			extensions.push(autocompletion());
		}

		if (readonly) {
			extensions.push(EditorState.readOnly.of(true));
		} else {
			// Add change listener
			extensions.push(EditorView.updateListener.of((update) => {
				if (update.docChanged) {
					dispatch('change', update.state.doc.toString());
				}
			}));
		}

		return extensions;
	}

	onMount(() => {
		const state = EditorState.create({
			doc: content,
			extensions: getExtensions()
		});

		view = new EditorView({
			state,
			parent: editorContainer
		});
	});

	onDestroy(() => {
		view?.destroy();
	});

	// Update content when prop changes
	$: if (view && content !== view.state.doc.toString()) {
		view.dispatch({
			changes: {
				from: 0,
				to: view.state.doc.length,
				insert: content
			}
		});
	}

	// Update highlight when prop changes
	$: if (view) {
		view.dispatch({
			effects: highlightEffect.of(highlightLines)
		});

		// Scroll to highlighted lines
		if (highlightLines) {
			try {
				const line = view.state.doc.line(highlightLines.from);
				view.dispatch({
					effects: EditorView.scrollIntoView(line.from, { y: 'start', yMargin: 100 })
				});
			} catch {
				// Line doesn't exist
			}
		}
	}

	export function clearHighlight() {
		if (view) {
			view.dispatch({
				effects: highlightEffect.of(null)
			});
		}
	}

	export function scrollToLine(lineNumber: number) {
		if (view) {
			try {
				const line = view.state.doc.line(lineNumber);
				view.dispatch({
					effects: EditorView.scrollIntoView(line.from, { y: 'start', yMargin: 100 })
				});
			} catch {
				// Line doesn't exist
			}
		}
	}
</script>

<div bind:this={editorContainer} class="h-full w-full overflow-auto"></div>

<style>
	div :global(.cm-editor) {
		height: 100%;
	}
	div :global(.cm-scroller) {
		overflow: auto;
	}
</style>
