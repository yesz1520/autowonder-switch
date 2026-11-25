# Documentation guidelines

English | [中文](README.zh.md)

<a id="pairs"></a>

## Document pairs

The repository opens in Chinese (`README.md`), with English in `README.en.md`. Both share `README.i18n.yaml`. Either language can be edited first. Under `docs/`, add, move and delete the following files together:

- `name.md` contains the English text.
- `name.zh.md` contains the Chinese text.
- `name.i18n.yaml` holds the Git blob hashes of the two editions last reviewed together.

Link to the other edition after the title and a blank line, as at the top of this page. Keep heading levels, explicit anchors, table shapes and code examples aligned. Repository links should lead readers to the corresponding page in their language.

<a id="voice"></a>

## Writing style

Start with the reader’s question. For a command, explain when to run it, which file it changes and how to tell it worked. If a sentence promises better reliability or easier maintenance, look for the missing fact. If the preceding sentence already explains the behavior, drop the claim. Keep design discussions separate; include a reason when it affects the reader’s choice. Cite the source revision or official documentation for third-party behavior.

For example, “Use `--record` to ensure the translations agree” suggests the command checks the translation. A maintainer needs the order of operations:

> After reviewing both editions, run `python3 scripts/check_docs.py --record docs/README.md`. The command writes the current hashes of both files to `docs/README.i18n.yaml`.

Use phrasing familiar to developers in the target language. Split or reorder sentences as needed, while preserving fields, commands, the object of each action and the order of operations. In the example, “after reviewing” tells the reader when to run the command. Keep that condition in the translation.

<a id="review"></a>

## Review and record

Before recording the pair, read the edited translation on its own. Can you tell who acts and under what conditions? Then compare the editions for requirements, optional behavior, failures and exceptions. For example, “cancel the current turn” and “cancel the entire conversation” lead to different implementations even if the examples match. Resolve those differences in the text first.

After review, run these commands from the repository root, substituting the English file you edited:

```sh
python3 scripts/check_docs.py --record docs/README.md
python3 scripts/check_docs.py
```

The first command updates the record for the named pair. The second checks all documentation pairs. It reports missing translations, stale records and differences in structure, link targets or examples.
