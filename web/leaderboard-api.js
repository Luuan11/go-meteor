let cachedLeaderboard = [];
let lastFetchTime = 0;
const CACHE_DURATION = 30000;
let gameSessionToken = null;
let lastScoreSaveTime = 0;
const MIN_SCORE_INTERVAL = 5000;

const API_URL = process.env.VERCEL_URL 
  ? `https://${process.env.VERCEL_URL}/api/leaderboard`
  : 'https://go-meteor-api.vercel.app/api/leaderboard';

async function loadLeaderboard() {
  const now = Date.now();
  
  if (cachedLeaderboard.length > 0 && (now - lastFetchTime) < CACHE_DURATION) {
    return cachedLeaderboard;
  }
  
  try {
    const response = await fetch(API_URL, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    
    if (data.leaderboard) {
      cachedLeaderboard = data.leaderboard;
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
  
  if (!gameSessionToken) {
    console.error('No valid session token');
    return false;
  }
  
  try {
    const response = await fetch(API_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        name: playerName,
        score: score,
        sessionToken: gameSessionToken,
        timestamp: Date.now()
      })
    });
    
    if (!response.ok) {
      const error = await response.json();
      console.error('Error saving score:', error);
      return false;
    }
    
    const result = await response.json();
    console.log('[Leaderboard] Score saved:', result);
    return true;
    
  } catch (error) {
    console.error('Error saving score:', error);
    return false;
  }
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
    const rankClass = index < 3 ? `rank-${rank}` : '';
    const displayName = truncateName(entry.name, 20);
    const displayScore = Math.min(entry.score, 999999);
    
    return `
      <div class="leaderboard-entry ${rankClass}">
        <div class="entry-left">
          <span class="entry-rank">${medal || rank}</span>
          <span class="entry-name">${escapeHtml(displayName)}</span>
        </div>
        <span class="entry-score">${displayScore.toLocaleString()}</span>
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

window.updateLeaderboard = async function(playerName, score) {
  console.log('[Leaderboard] Score saved to leaderboard:', playerName, '-', score, 'points');
  
  if (!gameSessionToken) {
    console.error('[Leaderboard] Invalid session - score rejected');
    return false;
  }
  
  const now = Date.now();
  if (now - lastScoreSaveTime < MIN_SCORE_INTERVAL) {
    console.error('[Leaderboard] Too many score updates - rate limited');
    return false;
  }
  
  lastScoreSaveTime = now;
  
  const success = await saveScore(playerName, score);
  if (success) {
    console.log('[Storage] Leaderboard saved to local storage');
    lastFetchTime = 0;
    const leaderboard = await loadLeaderboard();
    updateLeaderboardUI(leaderboard);
    gameSessionToken = null;
  }
  return success;
};

window.initGameSession = function() {
  const timestamp = Date.now();
  const random = Math.random().toString(36).substring(2, 15);
  gameSessionToken = `${timestamp}-${random}`;
  console.log('[Session] Game session initialized');
  return gameSessionToken;
};

document.addEventListener('DOMContentLoaded', async () => {
  console.log('[Leaderboard] Loading leaderboard...');
  const leaderboard = await loadLeaderboard();
  updateLeaderboardUI(leaderboard);
  
  setInterval(async () => {
    lastFetchTime = 0;
    const leaderboard = await loadLeaderboard();
    updateLeaderboardUI(leaderboard);
  }, 30000);
});
