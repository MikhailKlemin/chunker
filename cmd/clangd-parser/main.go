package main

import (
	"flag"
	"log"

	"clangd-parser/internal/lsp"
	"clangd-parser/internal/output"
	"clangd-parser/internal/parser"
)

func main() {
	compileDB := flag.String("compile-db", "/tmp", "Path to compile_commands.json directory")
	rootPath := flag.String("root", ".", "Root directory of the project")
	outputFile := flag.String("output", "chunks.json", "Output JSON file path")
	testFile := flag.String("test-file", "", "Single C++ file to test parsing")
	compact := flag.Bool("compact", false, "Write compact JSON (no indentation)")
	flag.Parse()

	log.Println("Clangd C++ Parser - Complete Pipeline")
	log.Println("======================================")

	// Step 1: Start LSP Client
	log.Println("\n→ Step 1: Starting clangd...")
	client, err := lsp.NewClient(*compileDB, *rootPath)
	if err != nil {
		log.Fatalf("❌ Failed to create LSP client: %v", err)
	}
	defer client.Close()

	info, _ := client.GetServerInfo()
	log.Printf("✓ %s", info)

	// Step 2: Discover C++ files
	log.Println("\n→ Step 2: Discovering C++ files...")
	files, err := parser.FindCppFiles(*rootPath)
	if err != nil {
		log.Fatalf("❌ Failed to find C++ files: %v", err)
	}

	// If test file specified, only parse that
	if *testFile != "" {
		files = []string{*testFile}
	}

	log.Printf("✓ Found %d C++ files to process", len(files))

	// Step 3: Parse symbols and create chunks
	log.Println("\n→ Step 3: Parsing document symbols...")

	var allChunks []parser.SemanticChunk
	successCount := 0
	errorCount := 0

	for i, file := range files {
		log.Printf("  [%d/%d] Processing %s", i+1, len(files), file)

		symbols, err := client.GetDocumentSymbols(file)
		if err != nil {
			log.Printf("  ⚠️  Warning: %v", err)
			errorCount++
			continue
		}

		chunks := parser.ConvertSymbolsToChunks(symbols, file)
		allChunks = append(allChunks, chunks...)
		successCount++

		if len(chunks) > 0 {
			log.Printf("  ✓ Created %d chunks", len(chunks))
		} else {
			log.Printf("  ℹ️  No extractable symbols found")
		}
	}

	log.Printf("\n✓ Processed %d files successfully (%d errors)", successCount, errorCount)
	log.Printf("✓ Total chunks created: %d", len(allChunks))

	// Step 4: Write output
	log.Println("\n→ Step 4: Writing output...")

	if *compact {
		err = output.WriteJSONCompact(allChunks, *outputFile)
	} else {
		err = output.WriteJSON(allChunks, *outputFile)
	}

	if err != nil {
		log.Fatalf("❌ Failed to write output: %v", err)
	}

	log.Printf("✓ Wrote output to: %s", *outputFile)

	// Show statistics
	stats := output.GetOutputStats(allChunks)
	log.Println("\n📊 Statistics:")
	log.Printf("  Total chunks: %d", stats["total_chunks"])
	log.Printf("  With docstrings: %d", stats["with_docstring"])

	log.Println("  By type:")
	byType := stats["by_type"].(map[string]int)
	for codeType, count := range byType {
		log.Printf("    %s: %d", codeType, count)
	}

	log.Println("\n✅ Complete! All steps finished successfully!")
}
