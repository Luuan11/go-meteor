let nameInputModal = null;
let nameInputField = null;
let nameInputCallback = null;
let eventListeners = [];
let isSubmitting = false;

function createNameInputModal() {
  const modal = document.createElement('div');
  modal.id = 'name-input-modal';
  modal.className = 'modal-overlay';
  modal.innerHTML = `
    <div class="modal-content">
      <div class="modal-header">
        <h2> TOP 10!</h2>
        <p>You made it to the leaderboard!</p>
      </div>
      <div class="modal-body">
        <label for="player-name-input">Enter your name:</label>
        <input 
          type="text" 
          id="player-name-input" 
          maxlength="15" 
          placeholder="Player Name"
          autocomplete="off"
        />
        <div class="char-counter">
          <span id="char-count">0</span>/15
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
  isSubmitting = false;
  
  const charCount = document.getElementById('char-count');
  const inputHandler = (e) => {
    charCount.textContent = e.target.value.length;
  };
  const keypressHandler = (e) => {
    if (e.key === 'Enter') {
      submitName();
    }
  };
  const keydownHandler = (e) => {
    if (e.key === 'Escape') {
      closeModal();
    }
  };
  const submitBtn = document.getElementById('submit-name-btn');
  const cancelBtn = document.getElementById('cancel-name-btn');
  const submitHandler = () => submitName();
  const cancelHandler = () => closeModal();
  const modalClickHandler = (e) => {
    if (e.target === modal) {
      closeModal();
    }
  };
  
  nameInputField.addEventListener('input', inputHandler);
  nameInputField.addEventListener('keypress', keypressHandler);
  nameInputField.addEventListener('keydown', keydownHandler);
  submitBtn.addEventListener('click', submitHandler);
  cancelBtn.addEventListener('click', cancelHandler);
  modal.addEventListener('click', modalClickHandler);
  
  eventListeners = [
    { element: nameInputField, event: 'input', handler: inputHandler },
    { element: nameInputField, event: 'keypress', handler: keypressHandler },
    { element: nameInputField, event: 'keydown', handler: keydownHandler },
    { element: submitBtn, event: 'click', handler: submitHandler },
    { element: cancelBtn, event: 'click', handler: cancelHandler },
    { element: modal, event: 'click', handler: modalClickHandler }
  ];
}

function submitName() {
  if (isSubmitting) return;
  isSubmitting = true;
  
  let name = nameInputField.value.trim();
  
  if (!name || name.length < 2) {
    nameInputField.classList.add('error');
    setTimeout(() => nameInputField.classList.remove('error'), 500);
    isSubmitting = false;
    return;
  }
  
  if (name.length > 15) {
    name = name.substring(0, 15);
  }
  
  const validName = name.replace(/[^a-zA-Z0-9\s\-_]/g, '').trim();
  if (!validName || validName.length < 2) {
    nameInputField.classList.add('error');
    setTimeout(() => nameInputField.classList.remove('error'), 500);
    isSubmitting = false;
    return;
  }
  
  if (nameInputCallback) {
    nameInputCallback(validName);
    nameInputCallback = null;
  }
  closeModal();
}

function closeModal() {
  if (nameInputCallback && !isSubmitting) {
    nameInputCallback('');
    nameInputCallback = null;
  }
  
  if (nameInputModal) {
    nameInputModal.classList.remove('show');
    
    eventListeners.forEach(({ element, event, handler }) => {
      if (element) {
        element.removeEventListener(event, handler);
      }
    });
    eventListeners = [];
    
    setTimeout(() => {
      if (nameInputModal && nameInputModal.parentNode) {
        nameInputModal.parentNode.removeChild(nameInputModal);
      }
      nameInputModal = null;
      nameInputField = null;
      isSubmitting = false;
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
