function decodeBase64URL(value) {
    const padding = (4 - (value.length % 4)) % 4;
    const base64 = value.replace(/-/g, '+').replace(/_/g, '/') + '='.repeat(padding);
    const binary = atob(base64);
    return Uint8Array.from(binary, character => character.charCodeAt(0)).buffer;
}

function encodeBase64URL(value) {
    const binary = Array.from(new Uint8Array(value), byte => String.fromCharCode(byte)).join('');
    return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

function creationOptionsFromJSON(options) {
    options.challenge = decodeBase64URL(options.challenge);
    options.user.id = decodeBase64URL(options.user.id);
    options.excludeCredentials = (options.excludeCredentials || []).map(credential => ({
        ...credential,
        id: decodeBase64URL(credential.id),
    }));
    return options;
}

function requestOptionsFromJSON(options) {
    options.challenge = decodeBase64URL(options.challenge);
    options.allowCredentials = (options.allowCredentials || []).map(credential => ({
        ...credential,
        id: decodeBase64URL(credential.id),
    }));
    return options;
}

async function responseJSON(response, message) {
    if (!response.ok) {
        const detail = await response.text();
        throw new Error(detail || message);
    }
    return response.json();
}

function jsonHeaders(form) {
    const headers = {'Content-Type': 'application/json'};
    const csrfField = form.querySelector('input[name="gorilla.csrf.Token"]');
    if (csrfField) {
        headers['X-CSRF-Token'] = csrfField.value;
    }
    return headers;
}

async function registerPasskey(form) {
	const name = form.elements.name.value.trim();
	const creation = await responseJSON(
		await fetch('/usr/passkeys/add/begin?name=' + encodeURIComponent(name)),
        'Failed to start passkey registration',
    );
    const credential = await navigator.credentials.create({
        publicKey: creationOptionsFromJSON(creation.publicKey),
    });
    if (!credential) {
        throw new Error('The authenticator did not create a credential');
    }

    const response = await fetch('/usr/passkeys/add/finish', {
        method: 'POST',
        headers: jsonHeaders(form),
        body: JSON.stringify({
            id: credential.id,
            rawId: encodeBase64URL(credential.rawId),
            type: credential.type,
            response: {
                attestationObject: encodeBase64URL(credential.response.attestationObject),
                clientDataJSON: encodeBase64URL(credential.response.clientDataJSON),
            },
        }),
    });
    if (!response.ok) {
        throw new Error((await response.text()) || 'Failed to finish passkey registration');
    }
    window.location.reload();
}

async function loginWithPasskey(form) {
    const username = form.elements.username.value;
    const request = await responseJSON(
        await fetch('/login/passkey/begin?username=' + encodeURIComponent(username)),
        'Passkey login is not available for this user',
    );
    const credential = await navigator.credentials.get({
        publicKey: requestOptionsFromJSON(request.publicKey),
    });
    if (!credential) {
        throw new Error('The authenticator did not return a credential');
    }

    const response = await fetch('/login/passkey/finish', {
        method: 'POST',
        headers: jsonHeaders(form),
        body: JSON.stringify({
            id: credential.id,
            rawId: encodeBase64URL(credential.rawId),
            type: credential.type,
            response: {
                authenticatorData: encodeBase64URL(credential.response.authenticatorData),
                clientDataJSON: encodeBase64URL(credential.response.clientDataJSON),
                signature: encodeBase64URL(credential.response.signature),
                userHandle: credential.response.userHandle ? encodeBase64URL(credential.response.userHandle) : null,
            },
        }),
    });
    if (!response.ok) {
        throw new Error((await response.text()) || 'Passkey login failed');
    }
    window.location.assign('/');
}

document.addEventListener('DOMContentLoaded', () => {
    const registrationForm = document.getElementById('register-passkey-form');
    if (registrationForm) {
        registrationForm.addEventListener('submit', async event => {
            event.preventDefault();
            try {
                await registerPasskey(registrationForm);
            } catch (error) {
                console.error(error);
                alert('Passkey registration failed: ' + error.message);
            }
        });
    }

    const loginForm = document.getElementById('passkey-login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', async event => {
            event.preventDefault();
            try {
                await loginWithPasskey(loginForm);
            } catch (error) {
                console.error(error);
                alert('Passkey login failed: ' + error.message);
            }
        });
    }
});
