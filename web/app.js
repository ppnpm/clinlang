// app.js — Wasm loader and UI logic

document.addEventListener("DOMContentLoaded", () => {
    const statusBadge = document.getElementById("app-status");
    const statusDot = statusBadge.querySelector(".dot");
    const inputArea = document.getElementById("cln-input");
    const outputArea = document.getElementById("soap-output");
    const alertsContainer = document.getElementById("alerts-container");
    const copyBtn = document.getElementById("btn-copy");

    // Load previously saved text from localStorage
    const savedText = localStorage.getItem("clinlang_autosave");
    if (savedText) {
        inputArea.value = savedText;
    }

    // Initialize WebAssembly
    const go = new Go(); // Provided by wasm_exec.js
    WebAssembly.instantiateStreaming(fetch("clinlang.wasm?v=3"), go.importObject).then((result) => {
        go.run(result.instance);
        
        // Update UI status
        statusDot.classList.remove("loading");
        statusBadge.innerHTML = '<span class="dot"></span> Engine Ready (Offline)';
        
        reloadCommands("general");
        
        // Run initial parse
        updateNote();
    }).catch(err => {
        console.error("Wasm load error:", err);
        statusDot.classList.remove("loading");
        statusBadge.innerHTML = `<span class="dot" style="background:red;"></span> Load Failed`;
        outputArea.textContent = "Error loading parsing engine. Ensure clinlang.wasm exists alongside index.html.";
    });

    let commandsList = [];

    function reloadCommands(profileStr) {
        if (typeof getAutocompleteCommands === 'function') {
            try {
                const cmdJson = getAutocompleteCommands(profileStr);
                commandsList = JSON.parse(cmdJson);
            } catch(e) {
                console.error("Failed to load autocomplete commands", e);
            }
        }
    }

    const autocompleteList = document.getElementById("autocomplete-list");
    let currentFocus = -1;
    let wordStartIdx = -1;
    let wordEndIdx = -1;

    // Handle input changes (Live Preview + Autocomplete)
    inputArea.addEventListener("input", (e) => {
        localStorage.setItem("clinlang_autosave", inputArea.value);
        updateNote();
        handleAutocomplete(e);
    });

    function handleAutocomplete(e) {
        let val = inputArea.value;
        let cursor = inputArea.selectionStart;
        
        // Find the start of the current word
        wordStartIdx = cursor;
        while (wordStartIdx > 0 && !/\n/.test(val[wordStartIdx - 1]) && !/\s/.test(val[wordStartIdx - 1])) {
            wordStartIdx--;
        }
        wordEndIdx = cursor;
        let word = val.substring(wordStartIdx, wordEndIdx).toLowerCase();
        
        // Find the start of the current line
        let lineStartIdx = cursor;
        while (lineStartIdx > 0 && !/\n/.test(val[lineStartIdx - 1])) {
            lineStartIdx--;
        }
        let currentLineUpToCursor = val.substring(lineStartIdx, cursor).trimStart().toLowerCase();

        closeAllLists();
        
        // Only show if word is at least 1 character and starts the line or starts a word
        if (!word) return false;

        let matches = [];

        // Check if we are inside an rx command context
        if (currentLineUpToCursor.startsWith("rx ") && typeof searchDrugs === 'function') {
            const rxPrefix = currentLineUpToCursor.substring(3).trimStart();
            if (rxPrefix.length > 0) {
                const drugResultsStr = searchDrugs(rxPrefix);
                try {
                    const drugResults = JSON.parse(drugResultsStr);
                    matches = drugResults.map(d => ({ cmd: d, desc: "Drug", isDrug: true }));
                } catch(e) {
                    console.error("Error parsing drug results", e);
                }
            }
        } else {
            matches = commandsList.filter(c => c.cmd.startsWith(word));
        }

        if (matches.length === 0) return false;


        currentFocus = 0; 

        matches.forEach((match, index) => {
            let li = document.createElement("li");
            li.className = "autocomplete-item" + (index === 0 ? " active" : "");
            li.innerHTML = `<div class="autocomplete-cmd">${match.cmd}</div><div class="autocomplete-desc">${match.desc}</div>`;
            li.addEventListener("click", function() {
                insertCompletion(match.cmd, match.isDrug);
            });
            autocompleteList.appendChild(li);
        });
        
        autocompleteList.classList.remove("hidden");
    }

    inputArea.addEventListener("keydown", function(e) {
        if (autocompleteList.classList.contains("hidden")) return;
        let items = autocompleteList.getElementsByTagName("li");
        
        if (e.key === "ArrowDown") {
            currentFocus++;
            addActive(items);
            e.preventDefault();
        } else if (e.key === "ArrowUp") {
            currentFocus--;
            addActive(items);
            e.preventDefault();
        } else if (e.key === "Enter" || e.key === "Tab") {
            e.preventDefault();
            if (currentFocus > -1 && items.length > 0) {
                items[currentFocus].click();
            }
        } else if (e.key === "Escape") {
            closeAllLists();
        }
    });

    function addActive(items) {
        if (!items) return false;
        removeActive(items);
        if (currentFocus >= items.length) currentFocus = 0;
        if (currentFocus < 0) currentFocus = (items.length - 1);
        items[currentFocus].classList.add("active");
        items[currentFocus].scrollIntoView({ block: "nearest" });
    }
    
    function removeActive(items) {
        for (let i = 0; i < items.length; i++) {
            items[i].classList.remove("active");
        }
    }

    function closeAllLists() {
        autocompleteList.innerHTML = "";
        autocompleteList.classList.add("hidden");
        currentFocus = -1;
    }

    function insertCompletion(cmd, isDrug = false) {
        let val = inputArea.value;
        if (isDrug) {
            let cursor = inputArea.selectionStart;
            let lineStartIdx = cursor;
            while (lineStartIdx > 0 && !/\n/.test(val[lineStartIdx - 1])) {
                lineStartIdx--;
            }
            let lineUpToCursor = val.substring(lineStartIdx, cursor);
            let rxIdx = lineUpToCursor.toLowerCase().lastIndexOf("rx ");
            if (rxIdx !== -1) {
                let replaceStart = lineStartIdx + rxIdx + 3;
                while (replaceStart < cursor && val[replaceStart] === ' ') {
                    replaceStart++;
                }
                inputArea.value = val.substring(0, replaceStart) + cmd + " " + val.substring(cursor);
                closeAllLists();
                inputArea.focus();
                inputArea.selectionStart = inputArea.selectionEnd = replaceStart + cmd.length + 1;
                updateNote();
                localStorage.setItem("clinlang_autosave", inputArea.value);
                return;
            }
        }

        inputArea.value = val.substring(0, wordStartIdx) + cmd + " " + val.substring(wordEndIdx);
        closeAllLists();
        inputArea.focus();
        inputArea.selectionStart = inputArea.selectionEnd = wordStartIdx + cmd.length + 1;
        updateNote();
        localStorage.setItem("clinlang_autosave", inputArea.value);
    }
    
    document.addEventListener("click", function (e) {
        if (e.target !== inputArea && e.target !== autocompleteList && !autocompleteList.contains(e.target)) {
            closeAllLists();
        }
    });

    // Copy to clipboard
    copyBtn.addEventListener("click", () => {
        const text = outputArea.textContent;
        navigator.clipboard.writeText(text).then(() => {
            const originalText = copyBtn.innerHTML;
            copyBtn.innerHTML = "Copied! ✓";
            copyBtn.style.backgroundColor = "var(--color-success)";
            setTimeout(() => {
                copyBtn.innerHTML = originalText;
                copyBtn.style.backgroundColor = "";
            }, 2000);
        });
    });

    // The core function executing logic inside the Wasm block
    function updateNote() {
        if (typeof parseClinLang !== 'function') return;

        const text = inputArea.value;
        if (!text.trim()) {
            outputArea.textContent = "";
            alertsContainer.classList.add("hidden");
            return;
        }

        // Call our Go engine!
        const jsonResultStr = parseClinLang(text);
        
        try {
            const result = JSON.parse(jsonResultStr);
            
            if (result.error) {
                outputArea.textContent = "Error: " + result.error;
                return;
            }

            // Sync autocomplete with current profile
            const currentProfile = result.json.profile || "general";
            reloadCommands(currentProfile);

            // Update output
            outputArea.textContent = result.soap || "No structured data parsed. Begin by typing a command like 'pt 45M'.";

            // Update Alerts
            renderAlerts(result.abnormal_flags || [], result.warnings || []);

        } catch (e) {
            outputArea.textContent = "Critical execution error parsing result: " + e.message;
        }
    }

    function renderAlerts(flags, warnings) {
        alertsContainer.innerHTML = "";
        
        if (flags.length === 0 && warnings.length === 0) {
            alertsContainer.classList.add("hidden");
            return;
        }
        
        alertsContainer.classList.remove("hidden");

        // Render Critical & Warning flags
        flags.forEach(flag => {
            const div = document.createElement("div");
            const isCritical = flag.severity === "critical";
            div.className = `alert-item ${isCritical ? 'alert-critical' : 'alert-warning'}`;
            
            const icon = isCritical ? '🔴' : '⚠️';
            div.innerHTML = `<span>${icon}</span> <span><strong>${flag.field} ${flag.value}</strong> — ${flag.message}</span>`;
            alertsContainer.appendChild(div);
        });

        // Render Syntax Warnings
        warnings.forEach(warn => {
            const div = document.createElement("div");
            div.className = `alert-item`;
            div.style.backgroundColor = "rgba(255,255,255,0.05)";
            div.style.border = "1px solid rgba(255,255,255,0.1)";
            
            div.innerHTML = `<span>💬</span> <span>Parser note: ${warn}</span>`;
            alertsContainer.appendChild(div);
        });
    }
});
