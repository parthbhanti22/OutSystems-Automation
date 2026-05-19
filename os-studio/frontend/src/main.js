import { GenerateApp } from '../wailsjs/go/main/App.js';

document.addEventListener('DOMContentLoaded', () => {
    const mermaidInput = document.getElementById('mermaidInput');
    const dropZone = document.getElementById('dropZone');
    const fileInput = document.getElementById('fileInput');
    const fileInfo = document.getElementById('fileInfo');
    const fileNameDisplay = document.getElementById('fileName');
    const removeFileBtn = document.getElementById('removeFileBtn');
    
    const generateBtn = document.getElementById('generateBtn');
    const btnText = generateBtn.querySelector('.btn-text');
    const spinner = generateBtn.querySelector('.spinner');
    const statusMessage = document.getElementById('statusMessage');

    let uploadedJson = null;

    // Drag and Drop Logic
    dropZone.addEventListener('dragover', (e) => {
        e.preventDefault();
        dropZone.classList.add('dragover');
    });

    dropZone.addEventListener('dragleave', () => {
        dropZone.classList.remove('dragover');
    });

    dropZone.addEventListener('drop', (e) => {
        e.preventDefault();
        dropZone.classList.remove('dragover');
        
        if (e.dataTransfer.files.length > 0) {
            handleFile(e.dataTransfer.files[0]);
        }
    });

    dropZone.addEventListener('click', () => {
        fileInput.click();
    });

    fileInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            handleFile(e.target.files[0]);
        }
    });

    removeFileBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        uploadedJson = null;
        fileInput.value = '';
        dropZone.classList.remove('hidden');
        fileInfo.classList.add('hidden');
    });

    function handleFile(file) {
        if (!file.name.endsWith('.json')) {
            setStatus("Please upload a valid JSON file.", "error");
            return;
        }

        const reader = new FileReader();
        reader.onload = (e) => {
            try {
                // Validate JSON
                JSON.parse(e.target.result);
                uploadedJson = e.target.result;
                fileNameDisplay.textContent = file.name;
                dropZone.classList.add('hidden');
                fileInfo.classList.remove('hidden');
                setStatus("");
            } catch (err) {
                setStatus("Invalid JSON format in file.", "error");
            }
        };
        reader.readAsText(file);
    }

    // Generate App Logic
    generateBtn.addEventListener('click', async () => {
        const mermaidData = mermaidInput.value.trim();
        
        if (!mermaidData && !uploadedJson) {
            setStatus("Please provide Mermaid code or a JSON schema.", "error");
            return;
        }

        // UI Loading State
        btnText.textContent = "Generating...";
        spinner.classList.remove('hidden');
        generateBtn.disabled = true;
        setStatus("Analyzing inputs and scaffolding application...", "");

        try {
            // Call the Go backend function
            // We pass an object with both potential inputs
            const payload = JSON.stringify({
                mermaid: mermaidData,
                json: uploadedJson || ""
            });
            
            const result = await GenerateApp(payload);
            setStatus(result, "success");
            
        } catch (err) {
            const msg = typeof err === 'string' ? err : (err.message || "An error occurred during generation.");
            setStatus(msg, "error");
        } finally {
            // Restore UI
            btnText.textContent = "Generate Application";
            spinner.classList.add('hidden');
            generateBtn.disabled = false;
        }
    });

    function setStatus(msg, type) {
        statusMessage.textContent = msg;
        statusMessage.className = 'status-message';
        if (type === 'error') statusMessage.classList.add('status-error');
        if (type === 'success') statusMessage.classList.add('status-success');
    }
});
