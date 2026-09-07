import tempfile
import unittest
from pathlib import Path

from rag_tutor import NO_EVIDENCE, LocalRAG, cosine_similarity, discover_markdown_files, safe_markdown_path, split_markdown


class FakeEmbedder:
    def __call__(self, text, task_type):
        lowered = text.lower()
        return [float("pedido" in lowered), float("coelho" in lowered), 0.1]


class ChunkTests(unittest.TestCase):
    def test_split_respects_size_and_overlap(self):
        chunks = split_markdown("um dois três quatro cinco seis", "README.md", chunk_size=16, overlap=5)
        self.assertGreater(len(chunks), 1)
        self.assertEqual([chunk.position for chunk in chunks], list(range(len(chunks))))
        self.assertTrue(all(chunk.source == "README.md" for chunk in chunks))
        self.assertIn(set(chunks[0].text.split()) & set(chunks[1].text.split()), ({"três"}, {"dois", "três"}))


class SearchTests(unittest.TestCase):
    def test_cosine_and_search_return_best_source(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "README.md").write_text("O pedido possui itens e status.", encoding="utf-8")
            (root / "DIAGNOSTICO.md").write_text("Um coelho aparece neste exemplo.", encoding="utf-8")
            rag = LocalRAG(root, FakeEmbedder())
            rag.build()
            results = rag.search("Como funciona o pedido?", min_score=0.5)
            self.assertEqual(results[0].chunk.source, "README.md")
            self.assertGreater(cosine_similarity([1, 0], [1, 0]), cosine_similarity([1, 0], [0, 1]))

    def test_no_evidence_is_explicit(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "README.md").write_text("Pedidos", encoding="utf-8")
            rag = LocalRAG(root, FakeEmbedder())
            rag.build()
            self.assertEqual(rag.context("assunto desconhecido", min_score=0.99), NO_EVIDENCE)


class FileProtectionTests(unittest.TestCase):
    def test_discovers_only_allowlisted_markdown(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "README.md").write_text("ok", encoding="utf-8")
            (root / "AGENTS.md").write_text("segredo", encoding="utf-8")
            (root / ".env").write_text("KEY=secret", encoding="utf-8")
            (root / "docs").mkdir()
            (root / "docs" / "guia.md").write_text("guia", encoding="utf-8")
            found = {path.relative_to(root).as_posix() for path in discover_markdown_files(root)}
            self.assertEqual(found, {"README.md", "docs/guia.md"})

    def test_rejects_outside_git_env_and_non_allowlisted_files(self):
        with tempfile.TemporaryDirectory() as directory, tempfile.TemporaryDirectory() as outside:
            root = Path(directory)
            protected = [root / ".env", root / ".git" / "config.md", root / "AGENTS.md", Path(outside) / "fora.md"]
            for path in protected:
                with self.subTest(path=path), self.assertRaises(ValueError):
                    safe_markdown_path(root, path)


if __name__ == "__main__":
    unittest.main()
