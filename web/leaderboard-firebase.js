let firebaseApp;
let database;
let leaderboardRef;
let cachedLeaderboard = [];
let lastFetchTime = 0;
const CACHE_DURATION = 30000;
let saveTimeout;
let gameSessionToken = null;
let lastScoreSaveTime = 0;
const MIN_SCORE_INTERVAL = 5000;

function initFirebase() {
  try {
    if (!window.firebaseConfig) {
      console.error('Firebase config not loaded! Include firebase-config.js before this script.');
      return;
    }
    
    Object.freeze(window.firebaseConfig);
    
    firebaseApp = firebase.initializeApp(window.firebaseConfig);
    database = firebase.database();
    leaderboardRef = database.ref('leaderboard');
    
    leaderboardRef.on('value', (snapshot) => {
      const data = snapshot.val();
      if (data) {
        cachedLeaderboard = Object.values(data)
          .sort((a, b) => b.score - a.score)
          .slice(0, 10);
        lastFetchTime = Date.now();
        updateLeaderboardUI(cachedLeaderboard);
      }
    });
  } catch (error) {
    console.error('Firebase initialization error:', error);
  }
}

async function loadLeaderboard() {
  const now = Date.now();
  
  if (cachedLeaderboard.length > 0 && (now - lastFetchTime) < CACHE_DURATION) {
    return cachedLeaderboard;
  }
  
  try {
    const snapshot = await leaderboardRef.orderByChild('score').limitToLast(10).once('value');
    const data = snapshot.val();
    
    if (data) {
      cachedLeaderboard = Object.values(data)
        .sort((a, b) => b.score - a.score);
      lastFetchTime = now;
      return cachedLeaderboard;
    }
    return [];
  } catch (error) {
    console.error('Error loading leaderboard:', error);
    return cachedLeaderboard;
  }
}

async function saveScore(playerName, score) {
  if (!playerName || playerName.length < 2 || playerName.length > 20) {
    console.error('Invalid player name');
    return false;
  }
  
  if (score < 0 || score > 999999) {
    console.error('Invalid score');
    return false;
  }
  
  clearTimeout(saveTimeout);
  
  return new Promise((resolve) => {
    saveTimeout = setTimeout(async () => {
      try {
        const newEntry = {
          name: playerName,
          score: score,
          timestamp: Date.now()
        };
        
        await leaderboardRef.push(newEntry);
        
        const snapshot = await leaderboardRef.once('value');
        const allEntries = snapshot.val();
        
        if (allEntries) {
          const entries = Object.entries(allEntries)
            .map(([key, value]) => ({key, ...value}))
            .sort((a, b) => b.score - a.score);
          
          if (entries.length > 10) {
            const toRemove = entries.slice(10);
            for (const entry of toRemove) {
              await leaderboardRef.child(entry.key).remove();
            }
          }
        }
        resolve(true);
      } catch (error) {
        console.error('Error saving score:', error);
        resolve(false);
      }
    }, 500);
  });
}

function updateLeaderboardUI(leaderboard) {
  const container = document.getElementById('leaderboard-entries');
  if (!container) return;
  
  if (!leaderboard || leaderboard.length === 0) {
    container.innerHTML = '<div class="empty-message">No scores yet. Be the first!</div>';
    return;
  }
  
  container.innerHTML = leaderboard.map((entry, index) => {
    const medal = index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : '';
    const rank = index + 1;
    const displayName = truncateName(entry.name, 20);
    const displayScore = Math.min(entry.score, 999999);
    return `
      <div class="leaderboard-entry ${index < 3 ? 'top-3' : ''}">
        <span class="rank">${medal || rank}</span>
        <span class="player-name">${escapeHtml(displayName)}</span>
        <span class="score">${displayScore.toLocaleString()}</span>
      </div>
    `;
  }).join('');
}

function truncateName(name, maxLength) {
  if (name.length <= maxLength) return name;
  return name.substring(0, maxLength - 3) + '...';
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

window.updateLeaderboard = async function(playerName, score, sessionToken) {
  if (!sessionToken || sessionToken !== gameSessionToken) {
    return false;
  }
  
  const now = Date.now();
  if (now - lastScoreSaveTime < MIN_SCORE_INTERVAL) {
    return false;
  }
  
  lastScoreSaveTime = now;
  gameSessionToken = null;
  
  const success = await saveScore(playerName, score);
  if (success) {
    const leaderboard = await loadLeaderboard();
    updateLeaderboardUI(leaderboard);
  }
  return success;
};

window.initGameSession = function() {
  gameSessionToken = Math.random().toString(36).substring(2) + Date.now().toString(36);
  return gameSessionToken;
};

document.addEventListener('DOMContentLoaded', async () => {
  initFirebase();
  const leaderboard = await loadLeaderboard();
  updateLeaderboardUI(leaderboard);
});
