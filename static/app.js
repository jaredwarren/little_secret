// Little Secret - Client Application

let socket = null;
let playerId = localStorage.getItem("little_secret_player_id") || "";
let roomCode = localStorage.getItem("little_secret_room_code") || "";
let playerName = localStorage.getItem("little_secret_player_name") || "";
let isHost = false;
let lobbyState = null;

// DOM Elements
const screens = {
    connect: document.getElementById("screen-connect"),
    lobby: document.getElementById("screen-lobby"),
    game: document.getElementById("screen-game"),
    reveal: document.getElementById("screen-reveal"),
    gameOver: document.getElementById("screen-game-over")
};

const logoButton = document.getElementById("logo-button");
const alertBanner = document.getElementById("alert-banner");
const alertMessage = document.getElementById("alert-message");
const btnCloseAlert = document.getElementById("btn-close-alert");

// Connection screen inputs
const inputNickname = document.getElementById("input-nickname");
const inputRoomCode = document.getElementById("input-room-code");
const btnCreateRoom = document.getElementById("btn-create-room");
const btnJoinRoom = document.getElementById("btn-join-room");

// Lobby screen elements
const lobbyCodeDisplay = document.getElementById("lobby-code-display");
const playerCountBadge = document.getElementById("player-count-badge");
const lobbyPlayersList = document.getElementById("lobby-players-list");
const btnStartGame = document.getElementById("btn-start-game");
const startGameWarning = document.getElementById("start-game-warning");
const selectGameMode = document.getElementById("select-game-mode");
const selectClueStyle = document.getElementById("select-clue-style");
const selectWordPack = document.getElementById("select-word-pack");
const selectWordNum = document.getElementById("select-word-num");
const btnToggleAdvanced = document.getElementById("btn-toggle-advanced");
const advancedSettings = document.getElementById("advanced-settings");
const inputConfusedCount = document.getElementById("input-confused-count");
const inputSpyCount = document.getElementById("input-spy-count");
const hostSettingsNotice = document.getElementById("host-settings-notice");

// Game screen elements
const gameModeBadge = document.getElementById("game-mode-badge");
const gameRoomBadge = document.getElementById("game-room-badge");
const gameStageTitle = document.getElementById("game-stage-title");
const roleCard = document.getElementById("role-card");
const roleImageContainer = document.getElementById("role-image-container");
const roleNameDisplay = document.getElementById("role-name-display");
const roleWordDisplay = document.getElementById("role-word-display");

// Panels
const panelOnline = document.getElementById("online-panels");
const panelClues = document.getElementById("panel-clues");
const inputClue = document.getElementById("input-clue");
const btnSubmitClue = document.getElementById("btn-submit-clue");
const cluesStatusList = document.getElementById("clues-status-list");

const panelDebate = document.getElementById("panel-debate");
const debateCluesGrid = document.getElementById("debate-clues-grid");
const hostDebateControls = document.getElementById("host-debate-controls");
const btnAdvanceVoting = document.getElementById("btn-advance-voting");

const panelVoting = document.getElementById("panel-voting");
const votingPanelTitle = document.getElementById("voting-panel-title");
const votingPanelInstructions = document.getElementById("voting-panel-instructions");
const votingPlayersGrid = document.getElementById("voting-players-grid");
const votingTallyStatus = document.getElementById("voting-tally-status");

const panelInPerson = document.getElementById("panel-in-person");
const btnInPersonReveal = document.getElementById("btn-in-person-reveal");

// Reveal screen elements
const eliminatedPlayerName = document.getElementById("eliminated-player-name");
const eliminatedPlayerStatus = document.getElementById("eliminated-player-status");
const normalRevealResult = document.getElementById("normal-reveal-result");
const revealRoleBadge = document.getElementById("reveal-role-badge");
const revealRoleDetails = document.getElementById("reveal-role-details");
const spyGuessBox = document.getElementById("spy-guess-box");
const spyGuessControls = document.getElementById("spy-guess-controls");
const spyGuessWaiting = document.getElementById("spy-guess-waiting");
const inputSpyGuess = document.getElementById("input-spy-guess");
const btnSubmitSpyGuess = document.getElementById("btn-submit-spy-guess");
const hostRevealControls = document.getElementById("host-reveal-controls");
const btnAdvanceRound = document.getElementById("btn-advance-round");
const revealVotingBreakdown = document.getElementById("reveal-voting-breakdown");
const revealVotesList = document.getElementById("reveal-votes-list");
const revealMyInfoBox = document.getElementById("reveal-my-info-box");
const revealMyRoleBadge = document.getElementById("reveal-my-role-badge");
const revealMyWordBadge = document.getElementById("reveal-my-word-badge");

// Game over screen elements
const gameOverTitle = document.getElementById("game-over-title");
const gameOverSubtitle = document.getElementById("game-over-subtitle");
const gameOverGoodWord = document.getElementById("game-over-good-word");
const gameOverConfusedWord = document.getElementById("game-over-confused-word");
const gameOverPlayersBody = document.getElementById("game-over-players-body");
const hostGameOverControls = document.getElementById("host-game-over-controls");
const btnReturnLobby = document.getElementById("btn-return-lobby");

// Pack Manager elements
const btnOpenPacks = document.getElementById("btn-open-packs");
const btnClosePacks = document.getElementById("btn-close-packs");
const modalPacks = document.getElementById("modal-packs");
const packsInstalledList = document.getElementById("packs-installed-list");
const inputPackName = document.getElementById("input-pack-name");
const textareaPackJson = document.getElementById("textarea-pack-json");
const btnUploadPack = document.getElementById("btn-upload-pack");
const btnToggleTemplate = document.getElementById("btn-toggle-template");
const packTemplateCode = document.getElementById("pack-template-code");

// Page Setup / Event Listeners
window.addEventListener("load", () => {
    // Check for debug/test query parameters to run multiple players on one machine
    const urlParams = new URLSearchParams(window.location.search);
    const debugPlayer = urlParams.get("player");
    const debugRoom = urlParams.get("room");
    
    if (debugPlayer) {
        playerName = debugPlayer;
        const storageKey = `little_secret_debug_id_${debugPlayer}`;
        let cachedDebugId = sessionStorage.getItem(storageKey);
        if (!cachedDebugId) {
            cachedDebugId = "debug_" + debugPlayer + "_" + Math.floor(Math.random() * 10000);
            sessionStorage.setItem(storageKey, cachedDebugId);
        }
        playerId = cachedDebugId;
        roomCode = debugRoom || localStorage.getItem("little_secret_room_code") || "";
        inputNickname.value = playerName;
        if (roomCode) {
            inputRoomCode.value = roomCode;
        }
    } else {
        // Fill nickname from storage if present
        if (playerName) {
            inputNickname.value = playerName;
        }
    }
    
    // Fill dynamic options for manual word number (1-21)
    for (let i = 1; i <= 21; i++) {
        const opt = document.createElement("option");
        opt.value = i.toString();
        opt.textContent = `Card #${i}`;
        selectWordNum.appendChild(opt);
    }

    // Connect WS if we have a room code already
    if (debugPlayer) {
        if (roomCode) {
            connectWebSocket("join");
        } else {
            connectWebSocket(); // waiting for them to click create
        }
    } else if (roomCode && playerId) {
        connectWebSocket();
    }
});

logoButton.addEventListener("click", () => {
    if (socket) {
        if (confirm("Disconnect and return to main screen?")) {
            localStorage.clear();
            location.reload();
        }
    } else {
        location.reload();
    }
});

btnCloseAlert.addEventListener("click", () => {
    alertBanner.classList.add("hidden");
});

btnCreateRoom.addEventListener("click", () => {
    const nick = inputNickname.value.trim();
    if (!nick) {
        showAlert("Nickname is required to create a room.");
        return;
    }
    playerName = nick;
    localStorage.setItem("little_secret_player_name", nick);
    connectWebSocket("create");
});

btnJoinRoom.addEventListener("click", () => {
    const nick = inputNickname.value.trim();
    const code = inputRoomCode.value.trim().toUpperCase();
    if (!nick) {
        showAlert("Nickname is required to join a room.");
        return;
    }
    if (!code) {
        showAlert("Room Code is required.");
        return;
    }
    playerName = nick;
    roomCode = code;
    localStorage.setItem("little_secret_player_name", nick);
    localStorage.setItem("little_secret_room_code", code);
    connectWebSocket("join");
});

// Settings Changes (Host only updates server)
const triggerSettingsUpdate = () => {
    if (!isHost || !socket) return;
    const config = {
        mode: selectGameMode.value,
        packName: selectWordPack.value,
        confusedCount: parseInt(inputConfusedCount.value) || -1,
        spyCount: parseInt(inputSpyCount.value) || -1,
        manualWordNum: parseInt(selectWordNum.value) || 0,
        sequentialClues: selectClueStyle.value === "true"
    };
    sendAction("configure_room", { config });
};

selectGameMode.addEventListener("change", triggerSettingsUpdate);
selectClueStyle.addEventListener("change", triggerSettingsUpdate);
selectWordPack.addEventListener("change", triggerSettingsUpdate);
selectWordNum.addEventListener("change", triggerSettingsUpdate);
inputConfusedCount.addEventListener("input", triggerSettingsUpdate);
inputSpyCount.addEventListener("input", triggerSettingsUpdate);

btnToggleAdvanced.addEventListener("click", () => {
    advancedSettings.classList.toggle("hidden");
    btnToggleAdvanced.querySelector(".chevron").innerHTML = advancedSettings.classList.contains("hidden") ? "&#9662;" : "&#9652;";
});

btnStartGame.addEventListener("click", () => {
    sendAction("start_game");
});

// Game Card Flipping
roleCard.addEventListener("click", () => {
    roleCard.classList.toggle("flipped");
});

// Action Submissions
btnSubmitClue.addEventListener("click", () => {
    const clue = inputClue.value.trim();
    if (!clue) return;
    sendAction("submit_clue", { clue });
    inputClue.value = "";
    btnSubmitClue.disabled = true;
    inputClue.disabled = true;
});

btnAdvanceVoting.addEventListener("click", () => {
    sendAction("call_vote");
});

btnInPersonReveal.addEventListener("click", () => {
    if (confirm("Are you sure you want to reveal your role? This will eliminate you.")) {
        sendAction("reveal_role");
    }
});

btnSubmitSpyGuess.addEventListener("click", () => {
    const guess = inputSpyGuess.value.trim();
    if (!guess) return;
    sendAction("guess_password", { guess });
});

btnAdvanceRound.addEventListener("click", () => {
    sendAction("next_round");
});

btnReturnLobby.addEventListener("click", () => {
    sendAction("restart_lobby");
});

// Pack Manager Modal listeners
btnOpenPacks.addEventListener("click", () => {
    modalPacks.classList.remove("hidden");
    loadPacksList();
});

btnClosePacks.addEventListener("click", () => {
    modalPacks.classList.add("hidden");
});

btnToggleTemplate.addEventListener("click", () => {
    packTemplateCode.classList.toggle("hidden");
    btnToggleTemplate.querySelector(".chevron").innerHTML = packTemplateCode.classList.contains("hidden") ? "&#9662;" : "&#9652;";
});

btnUploadPack.addEventListener("click", () => {
    const name = inputPackName.value.trim();
    const jsonStr = textareaPackJson.value.trim();
    if (!name || !jsonStr) {
        alert("Please fill in both name and JSON content.");
        return;
    }
    
    let words = [];
    try {
        const parsed = JSON.parse(jsonStr);
        words = parsed.words || parsed;
    } catch(e) {
        alert("Invalid JSON format.");
        return;
    }

    // Standardize pack body
    const body = { name, words };

    fetch("/api/packs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body)
    })
    .then(async res => {
        if (res.ok) {
            alert("Pack uploaded successfully!");
            inputPackName.value = "";
            textareaPackJson.value = "";
            loadPacksList();
            loadPacksDropdown(); // Refresh dropdown
        } else {
            const txt = await res.text();
            alert("Upload failed: " + txt);
        }
    })
    .catch(err => {
        alert("Error uploading pack: " + err);
    });
});

// Helper Functions
function showAlert(msg) {
    alertMessage.textContent = msg;
    alertBanner.classList.remove("hidden");
    // Auto scroll to alert
    alertBanner.scrollIntoView({ behavior: "smooth" });
}

function showScreen(screenId) {
    Object.keys(screens).forEach(key => {
        if (key === screenId) {
            screens[key].classList.add("active");
        } else {
            screens[key].classList.remove("active");
        }
    });
}

function sendAction(action, data = {}) {
    if (!socket || socket.readyState !== WebSocket.OPEN) return;
    const msg = {
        action,
        playerId,
        roomCode,
        playerName,
        ...data
    };
    socket.send(JSON.stringify(msg));
}

// WebSocket Connection
function connectWebSocket(intent = "") {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const url = `${protocol}//${window.location.host}/ws`;
    
    socket = new WebSocket(url);
    
    socket.onopen = () => {
        console.log("WebSocket connection established");
        alertBanner.classList.add("hidden");
        
        if (intent === "create") {
            sendAction("create_room");
        } else if (intent === "join") {
            sendAction("join_room");
        } else if (roomCode && playerId) {
            // Reconnecting
            sendAction("join_room");
        }
    };
    
    socket.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        switch (msg.type) {
            case "welcome":
                playerId = msg.welcome.playerId;
                roomCode = msg.welcome.roomCode;
                localStorage.setItem("little_secret_player_id", playerId);
                localStorage.setItem("little_secret_room_code", roomCode);
                break;
            case "state":
                lobbyState = msg.lobby;
                renderState();
                break;
            case "error":
                showAlert(msg.error);
                // Clear cache if room not found
                if (msg.error === "Room not found") {
                    localStorage.removeItem("little_secret_room_code");
                    roomCode = "";
                    showScreen("connect");
                }
                break;
        }
    };
    
    socket.onclose = () => {
        console.log("WebSocket connection closed");
        showAlert("Disconnected from server. Retrying connection in 3 seconds...");
        setTimeout(() => {
            if (roomCode && playerId) {
                connectWebSocket();
            }
        }, 3000);
    };

    socket.onerror = (err) => {
        console.error("WebSocket error:", err);
    };
}

// Render States
function renderState() {
    if (!lobbyState) return;

    const myPlayer = lobbyState.players[playerId];
    if (!myPlayer) {
        // We are no longer in this room or kicked
        localStorage.clear();
        location.reload();
        return;
    }

    isHost = myPlayer.isHost;

    // Load pack dropdown list if in lobby and dropdown is empty
    if (lobbyState.stage === "LOBBY") {
        loadPacksDropdown();
    }

    // Handle view screens depending on stage
    if (lobbyState.stage === "LOBBY") {
        showScreen("lobby");
        renderLobbyScreen(myPlayer);
    } else if (lobbyState.stage === "GAMEOVER") {
        showScreen("gameOver");
        renderGameOverScreen(myPlayer);
    } else if (lobbyState.stage === "REVEAL") {
        showScreen("reveal");
        renderRevealScreen(myPlayer);
    } else {
        // Game active stages: CLUES, DEBATE, VOTING, DEALT
        showScreen("game");
        renderGameScreen(myPlayer);
    }
}

// 1. Render Lobby
function renderLobbyScreen(myPlayer) {
    lobbyCodeDisplay.textContent = lobbyState.code;
    
    // Players List
    lobbyPlayersList.innerHTML = "";
    const playersArr = Object.values(lobbyState.players);
    playerCountBadge.textContent = `${playersArr.length} Players`;

    playersArr.forEach(p => {
        const li = document.createElement("li");
        li.className = p.isHost ? "host-item" : "";
        if (!p.connected) {
            li.classList.add("offline");
        }

        const nameWrapper = document.createElement("div");
        nameWrapper.className = "player-name-wrapper";
        nameWrapper.textContent = p.name;

        if (p.isHost) {
            const hBadge = document.createElement("span");
            hBadge.className = "host-badge";
            hBadge.textContent = "Host";
            nameWrapper.appendChild(hBadge);
        }

        const ind = document.createElement("span");
        ind.className = `conn-indicator ${p.connected ? '' : 'offline'}`;

        li.appendChild(nameWrapper);
        li.appendChild(ind);
        lobbyPlayersList.appendChild(li);
    });

    // Configure form settings
    const sequentialFormGroup = document.getElementById("form-group-sequential");
    if (isHost) {
        hostSettingsNotice.classList.add("hidden");
        selectGameMode.disabled = false;
        selectClueStyle.disabled = false;
        selectWordPack.disabled = false;
        selectWordNum.disabled = false;
        inputConfusedCount.disabled = false;
        inputSpyCount.disabled = false;
        
        btnStartGame.classList.remove("hidden");
        btnStartGame.disabled = playersArr.length < 4;
        startGameWarning.className = playersArr.length < 4 ? "warning-text" : "warning-text hidden";
    } else {
        hostSettingsNotice.classList.remove("hidden");
        selectGameMode.disabled = true;
        selectClueStyle.disabled = true;
        selectWordPack.disabled = true;
        selectWordNum.disabled = true;
        inputConfusedCount.disabled = true;
        inputSpyCount.disabled = true;
        
        btnStartGame.classList.add("hidden");
        startGameWarning.className = "warning-text hidden";
    }

    // Hide clue style in in-person mode
    if (selectGameMode.value === "IN_PERSON") {
        sequentialFormGroup.classList.add("hidden");
    } else {
        sequentialFormGroup.classList.remove("hidden");
    }

    // Sync input values (from state Config)
    selectGameMode.value = lobbyState.config.mode;
    selectClueStyle.value = lobbyState.config.sequentialClues ? "true" : "false";
    selectWordNum.value = lobbyState.config.manualWordNum.toString();
    inputConfusedCount.value = lobbyState.config.confusedCount === -1 ? "" : lobbyState.config.confusedCount.toString();
    inputSpyCount.value = lobbyState.config.spyCount === -1 ? "" : lobbyState.config.spyCount.toString();
    
    // Select pack safely
    if (selectWordPack.dataset.loaded === "true" && selectWordPack.querySelector(`option[value="${lobbyState.config.packName}"]`)) {
        selectWordPack.value = lobbyState.config.packName;
    }
}

// 2. Render Game Active
function renderGameScreen(myPlayer) {
    gameRoomBadge.textContent = `ROOM: ${lobbyState.code}`;
    gameModeBadge.textContent = lobbyState.config.mode;

    // Reset card class & flips
    roleCard.classList.remove("flipped");

    // Setup Front Card Text
    if (myPlayer.role && myPlayer.role.includes("Kitten")) {
        roleNameDisplay.textContent = "KITTEN";
        roleNameDisplay.className = "role-name good";
        roleWordDisplay.textContent = myPlayer.word || "UNKNOWN";
        roleCard.querySelector(".card-front").className = "card-face card-front good-kitten";
        
        roleImageContainer.className = "role-image-box kitten-theme";
        roleImageContainer.innerHTML = `<img src="cyberpunk_kitten.png" alt="Kitten Avatar">`;
    } else if (myPlayer.role === "Spy Pup") {
        roleNameDisplay.textContent = "SPY PUP";
        roleNameDisplay.className = "role-name spy";
        roleWordDisplay.textContent = "???";
        roleCard.querySelector(".card-front").className = "card-face card-front spy-pup";
        
        roleImageContainer.className = "role-image-box spy-theme";
        roleImageContainer.innerHTML = `<img src="cyberpunk_pup.png" alt="Spy Pup Avatar">`;
    } else {
        roleNameDisplay.textContent = "SPECTATOR";
        roleWordDisplay.textContent = "N/A";
    }

    // Stage visibility controls
    panelOnline.classList.add("hidden");
    panelInPerson.classList.add("hidden");
    panelClues.classList.add("hidden");
    panelDebate.classList.add("hidden");
    panelVoting.classList.add("hidden");

    if (lobbyState.config.mode === "IN_PERSON") {
        gameStageTitle.textContent = `VERBAL PLAY (Card #${lobbyState.activeWordNum})`;
        panelInPerson.classList.remove("hidden");
        
        // Hide card button if player is already eliminated
        if (myPlayer.isEliminated) {
            btnInPersonReveal.disabled = true;
            btnInPersonReveal.textContent = "Revealed & Eliminated";
        } else {
            btnInPersonReveal.disabled = false;
            btnInPersonReveal.textContent = "Reveal Role (If Eliminated)";
        }
    } else {
        // ONLINE MODE
        panelOnline.classList.remove("hidden");
        
        if (lobbyState.stage === "CLUES") {
            gameStageTitle.textContent = "SUBMIT CLUES";
            panelClues.classList.remove("hidden");

            if (lobbyState.config.sequentialClues) {
                const activePlayerID = lobbyState.turnOrder[lobbyState.currentTurnIdx];
                const isActivePlayer = activePlayerID === playerId;

                // Setup input state for sequential clues
                if (myPlayer.isEliminated) {
                    inputClue.disabled = true;
                    btnSubmitClue.disabled = true;
                    inputClue.placeholder = "You are eliminated...";
                } else if (myPlayer.clue !== "") {
                    inputClue.disabled = true;
                    btnSubmitClue.disabled = true;
                    inputClue.placeholder = `Clue sent: ${myPlayer.clue}`;
                } else if (isActivePlayer) {
                    inputClue.disabled = false;
                    btnSubmitClue.disabled = false;
                    inputClue.placeholder = "It's your turn! Type clue here...";
                } else {
                    inputClue.disabled = true;
                    btnSubmitClue.disabled = true;
                    const activePlayerName = lobbyState.players[activePlayerID]?.name || "another player";
                    inputClue.placeholder = `Waiting for ${activePlayerName} to submit clue...`;
                }

                // Fill sequential turn list
                cluesStatusList.innerHTML = "";
                lobbyState.turnOrder.forEach((tid, idx) => {
                    const p = lobbyState.players[tid];
                    if (!p) return;

                    const isCurrentTurn = idx === lobbyState.currentTurnIdx;
                    const badge = document.createElement("div");
                    
                    let statusText = "waiting";
                    if (p.clue !== "") {
                        statusText = p.clue; // Visible since it's sequential
                        badge.className = "status-badge submitted";
                    } else if (isCurrentTurn) {
                        statusText = "typing...";
                        badge.className = "status-badge active-turn-glow";
                    } else {
                        badge.className = "status-badge";
                    }
                    
                    badge.innerHTML = `<span>${p.name}</span><strong>${statusText}</strong>`;
                    cluesStatusList.appendChild(badge);
                });
            } else {
                // Setup input state for simultaneous clues
                if (myPlayer.isEliminated) {
                    inputClue.disabled = true;
                    btnSubmitClue.disabled = true;
                    inputClue.placeholder = "You are eliminated...";
                } else if (myPlayer.clue !== "") {
                    inputClue.disabled = true;
                    btnSubmitClue.disabled = true;
                    inputClue.placeholder = `Clue sent: ${myPlayer.clue}`;
                } else {
                    inputClue.disabled = false;
                    btnSubmitClue.disabled = false;
                    inputClue.placeholder = "Type clue here...";
                }

                // Fill submission grid
                cluesStatusList.innerHTML = "";
                Object.values(lobbyState.players).forEach(p => {
                    if (p.isEliminated) return;
                    const badge = document.createElement("div");
                    badge.className = p.clue !== "" ? "status-badge submitted" : "status-badge";
                    badge.innerHTML = `<span>${p.name}</span><span>${p.clue !== "" ? "✓" : "..."}</span>`;
                    cluesStatusList.appendChild(badge);
                });
            }

        } else if (lobbyState.stage === "DEBATE") {
            gameStageTitle.textContent = "DEBATE & DISCUSS";
            panelDebate.classList.remove("hidden");

            // Fill clues grid
            debateCluesGrid.innerHTML = "";
            Object.values(lobbyState.players).forEach(p => {
                const card = document.createElement("div");
                card.className = p.isEliminated ? "clue-card eliminated-card" : "clue-card";
                card.innerHTML = `
                    <span class="player-name">${p.name} ${p.isEliminated ? '(ELIMINATED)' : ''}</span>
                    <span class="clue-word">${p.isEliminated ? '—' : (p.clue || 'no clue')}</span>
                `;
                debateCluesGrid.appendChild(card);
            });

            // Host Call Vote button
            if (isHost) {
                hostDebateControls.classList.remove("hidden");
            } else {
                hostDebateControls.classList.add("hidden");
            }

        } else if (lobbyState.stage === "VOTING") {
            gameStageTitle.textContent = "VOTING STAGE";
            panelVoting.classList.remove("hidden");

            if (lobbyState.isTieVote) {
                votingPanelTitle.textContent = "Tie-Breaker Vote!";
                const tieNames = lobbyState.tiePlayers.map(tid => lobbyState.players[tid]?.name).join(" & ");
                votingPanelInstructions.textContent = `There is a tie between: ${tieNames}! Cast a new vote to break it.`;
            } else {
                votingPanelTitle.textContent = "Elimination Vote";
                votingPanelInstructions.textContent = "Cast your vote to eliminate the player you suspect the most!";
            }

            // Render voting target cards
            votingPlayersGrid.innerHTML = "";
            Object.values(lobbyState.players).forEach(p => {
                // Cannot vote for eliminated players or yourself
                if (p.isEliminated || p.id === playerId) return;
                
                // If it is a tie vote, we can only vote for players who are tied!
                if (lobbyState.isTieVote && !lobbyState.tiePlayers.includes(p.id)) return;

                const vCard = document.createElement("div");
                const isMyVoteTarget = myPlayer.vote === p.id;
                vCard.className = isMyVoteTarget ? "vote-target-card voted-for" : "vote-target-card";
                if (myPlayer.isEliminated) {
                    vCard.classList.add("disabled");
                }

                vCard.innerHTML = `
                    <div class="target-info">
                        <span class="target-name">${p.name}</span>
                        <span class="target-clue">${p.clue ? `Clue: ${p.clue}` : 'No clue'}</span>
                    </div>
                    <button class="btn-vote">${isMyVoteTarget ? 'Voted' : 'Vote'}</button>
                `;

                // Handle click vote
                if (!myPlayer.isEliminated) {
                    vCard.querySelector("button").addEventListener("click", () => {
                        sendAction("submit_vote", { targetId: p.id });
                    });
                }

                votingPlayersGrid.appendChild(vCard);
            });

            // Render tally list of who has voted
            votingTallyStatus.innerHTML = "";
            Object.values(lobbyState.players).forEach(p => {
                if (p.isEliminated) return;
                
                const tBadge = document.createElement("span");
                tBadge.className = p.vote !== "" ? "tally-badge voted" : "tally-badge";
                tBadge.textContent = `${p.name} ${p.vote !== "" ? '✓' : '...'}`;
                votingTallyStatus.appendChild(tBadge);
            });
        }
    }
}

// 3. Render Reveal
function renderRevealScreen(myPlayer) {
    const elimId = lobbyState.eliminatedThisTurn && lobbyState.eliminatedThisTurn.length > 0 ? lobbyState.eliminatedThisTurn[0] : null;
    const elimPlayer = elimId ? lobbyState.players[elimId] : null;

    if (!elimPlayer) {
        eliminatedPlayerName.textContent = "No one";
        eliminatedPlayerStatus.textContent = "was eliminated this round.";
        normalRevealResult.classList.add("hidden");
        spyGuessBox.classList.add("hidden");
    } else {
        eliminatedPlayerName.textContent = elimPlayer.name;
        eliminatedPlayerStatus.textContent = "was voted out! Checking role...";

        if (elimPlayer.role === "Spy Pup") {
            normalRevealResult.classList.add("hidden");
            spyGuessBox.classList.remove("hidden");

            // Update badge text with emoji
            const spyRoleBadge = spyGuessBox.querySelector(".reveal-role-badge");
            if (spyRoleBadge) {
                spyRoleBadge.textContent = "🐶 SPY PUP";
            }

            // Spy guess interaction
            if (elimId === playerId) {
                // If I am the Spy Pup
                spyGuessControls.classList.remove("hidden");
                spyGuessWaiting.classList.add("hidden");
                inputSpyGuess.disabled = false;
                btnSubmitSpyGuess.disabled = false;
            } else {
                spyGuessControls.classList.add("hidden");
                spyGuessWaiting.classList.remove("hidden");
                spyGuessWaiting.textContent = `Waiting for Spy Pup (${elimPlayer.name}) to enter their password guess...`;
            }
        } else {
            // Normal Kitten Reveal
            spyGuessBox.classList.add("hidden");
            normalRevealResult.classList.remove("hidden");

            if (elimPlayer.role === "Good Kitten") {
                revealRoleBadge.textContent = "🐱 GOOD KITTEN";
                revealRoleBadge.className = "reveal-role-badge";
                revealRoleDetails.textContent = `Their secret word matches the correct password: "${elimPlayer.word}"`;
            } else {
                revealRoleBadge.textContent = "😿 CONFUSED KITTEN";
                revealRoleBadge.className = "reveal-role-badge confused";
                revealRoleDetails.textContent = `They did not know they were confused! Their secret word was: "${elimPlayer.word}"`;
            }
        }
    }

    // Render Voting Breakdown (Online Mode only)
    if (lobbyState.config.mode === "ONLINE") {
        revealVotingBreakdown.classList.remove("hidden");
        
        // Calculate vote counts first to display in parentheses
        const voteCounts = {};
        Object.values(lobbyState.players).forEach(p => {
            if (p.vote && lobbyState.players[p.vote]) {
                voteCounts[p.vote] = (voteCounts[p.vote] || 0) + 1;
            }
        });

        // Filter and sort players who were active voters in this round
        const activeVoters = Object.values(lobbyState.players)
            .filter(p => !p.isEliminated || (lobbyState.eliminatedThisTurn && lobbyState.eliminatedThisTurn.includes(p.id)))
            .sort((a, b) => a.name.localeCompare(b.name));

        revealVotesList.innerHTML = "";
        activeVoters.forEach(p => {
            const li = document.createElement("li");

            const voterSpan = document.createElement("span");
            voterSpan.className = "voter";
            voterSpan.textContent = p.name;

            const arrowSpan = document.createElement("span");
            arrowSpan.className = "arrow";
            arrowSpan.textContent = "➔";

            const targetSpan = document.createElement("span");
            targetSpan.className = "target";

            if (p.vote && lobbyState.players[p.vote]) {
                const targetPlayer = lobbyState.players[p.vote];
                const count = voteCounts[p.vote] || 0;
                targetSpan.textContent = `${targetPlayer.name} (${count} ${count === 1 ? 'vote' : 'votes'})`;

                if (lobbyState.eliminatedThisTurn && lobbyState.eliminatedThisTurn.includes(p.vote)) {
                    targetSpan.classList.add("eliminated-target");
                }
            } else {
                targetSpan.textContent = "No Vote / Skipped";
                targetSpan.classList.add("skipped-vote");
            }

            li.appendChild(voterSpan);
            li.appendChild(arrowSpan);
            li.appendChild(targetSpan);
            revealVotesList.appendChild(li);
        });
    } else {
        revealVotingBreakdown.classList.add("hidden");
    }

    // Render My Info Reminder
    if (myPlayer) {
        revealMyInfoBox.classList.remove("hidden");
        
        let displayRole = "🐱 KITTEN";
        let roleClass = "my-role-badge";
        
        if (myPlayer.role === "Spy Pup") {
            displayRole = "🐶 SPY PUP";
            roleClass = "my-role-badge spy";
        } else if (myPlayer.role === "Good Kitten") {
            displayRole = "🐱 GOOD KITTEN";
            roleClass = "my-role-badge";
        } else if (myPlayer.role === "Confused Kitten") {
            displayRole = "😿 CONFUSED KITTEN";
            roleClass = "my-role-badge confused";
        }
        
        revealMyRoleBadge.textContent = displayRole;
        revealMyRoleBadge.className = roleClass;
        
        if (myPlayer.role === "Spy Pup") {
            revealMyWordBadge.textContent = "Word: None";
        } else {
            revealMyWordBadge.textContent = `Word: "${myPlayer.word}"`;
        }
    } else {
        revealMyInfoBox.classList.add("hidden");
    }

    // Host controls to advance
    if (isHost) {
        hostRevealControls.classList.remove("hidden");
    } else {
        hostRevealControls.classList.add("hidden");
    }
}

// 4. Render Game Over
function renderGameOverScreen(myPlayer) {
    gameOverGoodWord.textContent = lobbyState.goodWord;
    gameOverConfusedWord.textContent = lobbyState.confusedWord;

    // Set Winner Headline
    if (lobbyState.winner === "KITTENS") {
        gameOverTitle.textContent = "KITTENS WIN!";
        gameOverTitle.className = "highlight";
        gameOverSubtitle.textContent = "All Spy Pups and Confused Kittens were successfully rooted out!";
    } else if (lobbyState.winner === "SPY_PUP") {
        gameOverTitle.textContent = "SPY PUP WINS!";
        gameOverTitle.className = "highlight text-orange"; // style tweak done programmatically or class
        gameOverSubtitle.textContent = "The Spy Pup guessed the correct password or remained undetected!";
    } else if (lobbyState.winner === "CONFUSED_KITTENS") {
        gameOverTitle.textContent = "CONFUSED KITTENS WIN!";
        gameOverTitle.className = "highlight";
        gameOverSubtitle.textContent = "The Confused Kittens successfully bluffed and remained undetected!";
    } else {
        gameOverTitle.textContent = "GAME OVER";
        gameOverSubtitle.textContent = "The round has ended.";
    }

    // Fill Roles Summary Table
    gameOverPlayersBody.innerHTML = "";
    Object.values(lobbyState.players).forEach(p => {
        const tr = document.createElement("tr");
        
        let roleClass = "good";
        if (p.role === "Spy Pup") roleClass = "spy";
        else if (p.role === "Confused Kitten") roleClass = "confused";

        tr.innerHTML = `
            <td><strong>${p.name}</strong> ${p.isEliminated ? '<span style="opacity:0.5;font-size:0.8rem;">(Eliminated)</span>' : ''}</td>
            <td><span class="role-name ${roleClass}" style="font-size:0.9rem;">${p.role}</span></td>
            <td><code>${p.word || '—'}</code></td>
        `;
        gameOverPlayersBody.appendChild(tr);
    });

    if (isHost) {
        hostGameOverControls.classList.remove("hidden");
    } else {
        hostGameOverControls.classList.add("hidden");
    }
}

// Packs API Requests
function loadPacksList() {
    packsInstalledList.innerHTML = "<li>Loading packs...</li>";
    fetch("/api/packs")
        .then(res => res.json())
        .then(data => {
            packsInstalledList.innerHTML = "";
            data.forEach(name => {
                const li = document.createElement("li");
                li.textContent = name;
                packsInstalledList.appendChild(li);
            });
        })
        .catch(err => {
            packsInstalledList.innerHTML = `<li style="color:var(--neon-danger)">Error: ${err}</li>`;
        });
}

function loadPacksDropdown() {
    if (selectWordPack.dataset.loaded === "true") return; // Only load once per lobby session
    
    fetch("/api/packs")
        .then(res => res.json())
        .then(data => {
            selectWordPack.innerHTML = "";
            data.forEach(name => {
                const opt = document.createElement("option");
                opt.value = name;
                opt.textContent = name;
                selectWordPack.appendChild(opt);
            });
            selectWordPack.dataset.loaded = "true";
            
            // Sync again once loaded
            if (lobbyState && lobbyState.config) {
                selectWordPack.value = lobbyState.config.packName;
            }
        })
        .catch(err => {
            console.error("failed to load packs list dropdown:", err);
        });
}
