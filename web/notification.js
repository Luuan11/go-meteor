class NotificationManager {
  constructor() {
    this.container = null;
    this.defaultDuration = 3000;
    this.animationDelay = 10;
    this.removeDelay = 300;
    this.init();
  }

  init() {
    if (!document.getElementById('notification-container')) {
      this.container = document.createElement('div');
      this.container.id = 'notification-container';
      this.container.className = 'notification-container';
      document.body.appendChild(this.container);
    } else {
      this.container = document.getElementById('notification-container');
    }
  }

  show(message, type = 'info', duration = this.defaultDuration) {
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    
    const icon = this.getIcon(type);
    notification.innerHTML = `
      <span class="notification-icon">${icon}</span>
      <span class="notification-message">${this.escapeHtml(message)}</span>
    `;
    
    this.container.appendChild(notification);
    
    setTimeout(() => notification.classList.add('show'), this.animationDelay);
    
    setTimeout(() => {
      notification.classList.remove('show');
      setTimeout(() => {
        if (notification.parentNode) {
          notification.parentNode.removeChild(notification);
        }
      }, this.removeDelay);
    }, duration);
  }

  getIcon(type) {
    const icons = {
      success: '✓',
      error: '✕',
      warning: '⚠',
      info: 'ℹ'
    };
    return icons[type] || icons.info;
  }

  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  success(message, duration) {
    this.show(message, 'success', duration);
  }

  error(message, duration) {
    this.show(message, 'error', duration);
  }

  warning(message, duration) {
    this.show(message, 'warning', duration);
  }

  info(message, duration) {
    this.show(message, 'info', duration);
  }
}

window.notificationManager = new NotificationManager();
