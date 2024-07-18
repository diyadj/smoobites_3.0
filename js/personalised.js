document.addEventListener("DOMContentLoaded", function () {
    fetch('/session-info')
      .then(response => response.json())
      .then(data => {
        const userGreeting = document.getElementById('user-greeting');
        const logoutButton = document.getElementById('logout-button');

        if (data.user) {
          userGreeting.innerHTML = `${data.user}`;
          logoutButton.style.display = 'block';
          document.getElementById('SOELink').setAttribute('href', 'schools.html');
          document.getElementById('SCISLink').setAttribute('href', 'schools.html');
          document.getElementById('connexLink').setAttribute('href', 'schools.html');
          document.getElementById('orderNowLink').setAttribute('href', 'schools.html');
        } else {
          userGreeting.innerHTML = '<a href="login.html" role="button">Login</a>';
          logoutButton.style.display = 'none';
          document.getElementById('SOELink').setAttribute('href', '/login.html');
          document.getElementById('SCISLink').setAttribute('href', '/login.html');
          document.getElementById('connexLink').setAttribute('href', '/login.html');
          document.getElementById('orderNowLink').setAttribute('href', '/login.html');
        }
      });

    document.getElementById('logout-button').addEventListener('click', function () {
      fetch('/logout')
        .then(response => response.json())
        .then(data => {
          if (data.status === 'logged out') {
            const userGreeting = document.getElementById('user-greeting');
            const logoutButton = document.getElementById('logout-button');
            userGreeting.innerHTML = '<a href="login.html" role="button">Login</a>';
            logoutButton.style.display = 'none';
          }
        });
    });
  });