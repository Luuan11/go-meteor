let nameInputModal = null;
let nameInputField = null;
let nameInputCallback = null;

function createNameInputModal() {
  const modal = document.createElement('div');
  modal.id = 'name-input-modal';
  modal.className = 'modal-overlay';
  modal.innerHTML = `
    <div class="modal-content">
      <div class="modal-header">
        <h2>🏆 TOP 10!</h2>
        <p>You made it to the leaderboard!</p>
      </div>
      <div class="modal-body">
        <label for="player-name-input">Enter your name:</label>
        <input 
          type="text" 
          id="player-name-input" 
          maxlength="20" 
          placeholder="Player Name"
          autocomplete="off"
        />
        <div class="char-counter">
          <span id="char-count">0</span>/20
        </div>
      </div>
      <div class="modal-footer">
        <button id="submit-name-btn" class="btn-primary">Submit</button>
        <button id="cancel-name-btn" class="btn-secondary">Skip</button>
      </div>
    </div>
  `;
  
  document.body.appendChild(modal);
  nameInputModal = modal;
  nameInputField = document.getElementById('player-name-input');
  
  const charCount = document.getElementById('char-count');
  nameInputField.addEventListener('input', (e) => {
    charCount.textContent = e.target.value.length;
  });
  
  nameInputField.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
      submitName();
    }
  });
  
  document.getElementById('submit-name-btn').addEventListener('click', submitName);
  document.getElementById('cancel-name-btn').addEventListener('click', closeModal);
  
  modal.addEventListener('click', (e) => {
    if (e.target === modal) {
      closeModal();
    }
  });
}

function submitName() {
  const name = nameInputField.value.trim();
  if (name.length < 2) {
    nameInputField.classList.add('error');
    setTimeout(() => nameInputField.classList.remove('error'), 500);
    return;
  }
  
  if (nameInputCallback) {
    nameInputCallback(name);
  }
  closeModal();
}

function closeModal() {
  if (nameInputModal) {
    nameInputModal.classList.remove('show');
    setTimeout(() => {
      if (nameInputModal && nameInputModal.parentNode) {
        nameInputModal.parentNode.removeChild(nameInputModal);
      }
      nameInputModal = null;
      nameInputField = null;
      nameInputCallback = null;
    }, 300);
  }
}

window.showNameInputModal = function(callback) {
  if (!nameInputModal) {
    createNameInputModal();
  }
  
  nameInputCallback = callback;
  nameInputField.value = '';
  document.getElementById('char-count').textContent = '0';
  
  setTimeout(() => {
    nameInputModal.classList.add('show');
    nameInputField.focus();
  }, 10);
};
