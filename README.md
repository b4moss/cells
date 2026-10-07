# b4moss / cells

[Pydio Cells](https://github.com/pydio/cells) の **b4moss フォーク**です。継続利用とローカル改善のために保守しています。

b4moss が継続利用のために保守しているサードパーティフォークは、現時点で次の 2 つです。

- [`b4moss/gitea`](https://github.com/b4moss/gitea)
- [`b4moss/cells`](https://github.com/b4moss/cells)（本リポジトリ）

製品の機能・インストール・開発手順などの公式情報は、上流の README を参照してください。

- 上流リポジトリ: [pydio/cells](https://github.com/pydio/cells)
- 上流 README: [pydio/cells README](https://github.com/pydio/cells/blob/main/README.md)

## バージョン付け（リリースタグ）

b4moss がこのフォークをリリースするときのタグ形式は、次のとおりです。

```text
{original_version}-b4m{our_version}
```

例（形式の説明のみ。実在するリリースタグを示すものではありません）: `5.1.0-b4m1`

| 部分 | 意味 |
| --- | --- |
| `{original_version}` | ベースにした上流バージョン |
| `{our_version}` | このフォーク系列における b4moss 側のリリースカウンタ（単調増加） |

`{our_version}` は SemVer（major.minor.patch）ではありません。`b4m` 以降に major.minor.patch 形式を要求しません。

### タグと追従ブランチ

- **追従・追跡ブランチ**: 上流のベースラインを追うためのブランチ（例: 特定の上流版に追従する作業ブランチ）
- **リリースタグ**: b4moss としてのリリース時点を示すタグ（上記の `{original_version}-b4m{our_version}`）

ブランチとタグは役割を分けて扱います。追従ブランチ名とリリースタグを混同しないでください。

### 既存タグについて

本リポジトリには、上流と同様の見た目のタグ（例: `v5.1.0`）が残っている場合があります。上記の `{original_version}-b4m{our_version}` は **今後の目標スキーム**です。既存タグの付け直しや履歴の書き換えは行いません。
