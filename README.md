<div align="center">
  <a href="https://openline.ai">
    <img
      src="https://raw.githubusercontent.com/openline-ai/openline-voice/otter/.github/TeamHero.svg"
      alt="Openline Logo"
      height="64"
    />
  </a>
  <br />
  <p>
    <h3>
      <b>
        Openline Voice
      </b>
    </h3>
  </p>
  <p>
    Voice network and app plugin for the Openline ecosystem
  </p>
  <p>

[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen?logo=github)][repo] 
[![license](https://img.shields.io/badge/license-Apache%202-blue)][apache2] 
[![stars](https://img.shields.io/github/stars/openline-ai/openline-voice?style=social)][repo] 
[![twitter](https://img.shields.io/twitter/follow/openlineAI?style=social)][twitter] 
[![slack](https://img.shields.io/badge/slack-community-blueviolet.svg?logo=slack)][slack]

  </p>
  <p>
    <sub>
      Built with ❤︎ by the
      <a href="https://openline.ai">
        Openline
      </a>
      community!
    </sub>
  </p>
</div>


## 👋 Overview
![Octavian Tails On The Phone](images/otter_phone.jpeg)

The Openline Voice platform allows WebRTC to PSTN calling

## 🚀 Installation

Download using the following command

### set up a k3d environment

* This process has been tested onUbuntu 20.04 the install process may need to be adapted for other platforms
* If you use codespaces, be sure to use the 4 core environment
* The voice network can not run on arm64, x86_64 must be used


#### Run the deployment script

```
curl http://openline.sh/install.sh | sh
openline dev start -v voice
```

after the script completes you can validate the status of the setup by running
```
openline dev ping
```

if you are not running the minikube on your local machine and the minikube is behind a nat you will probably need to install the turn server to have audio
```
docker run -d --network=host coturn/coturn -v -z -n -X $(hostname -I|cut -f1 -d ' ') -L 0.0.0.0 --min-port=10000 --max-port=20000
```


## 🙌 Features

TBD

## 🤝 Resources

- For help, feature requests, or chat with fellow Openline enthusiasts, check out our [slack community][slack]!
- Our [docs site][docs] has references for developer functionality, including the Graph API

## 👩‍💻 Codebase

### Technologies

Here's a list of the big technologies that we use:

- TBD

### Folder structure

```sh
openline-voice/
├── architecture            # Architectural documentation
├── deployment              
│   ├── infra               # Infrastructure-as-code
│   └── scripts             # Deployment scripts
└── packages
    ├── apps                # Front end web applications
    │   ├── voice-plugin    # customerOS data explorer
    ├── core                # Shared core libraries
    └── server              # Back end database & API server
```


## 💪 Contributions

- We love contributions big or small!  Please check out our [guide on how to get started][contributions].
- Not sure where to start?  [Book a free, no-pressure, no-commitment call][call] with the team to discuss the best way to get involved.

## ✨ Contributors

A massive thank you goes out to all these wonderful people ([emoji key][emoji]):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->


<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

## 🪪 License

- This repo is licensed under [Apache 2.0][apache2], with the exception of the ee directory (if applicable).
- Premium features (contained in the ee directory) require an Openline Enterprise license.  See our [pricing page][pricing] for more details.


[apache2]: https://www.apache.org/licenses/LICENSE-2.0
[call]: https://meetings-eu1.hubspot.com/matt2/customer-demos
[contributions]: https://github.com/openline-ai/community/blob/main/README.md
[docs]: https://openline.ai
[emoji]: https://allcontributors.org/docs/en/emoji-key
[pricing]: https://openline.ai/pricing
[repo]: https://github.com/openline-ai/openline-voice/
[slack]: https://join.slack.com/t/openline-ai/shared_invite/zt-1i6umaw6c-aaap4VwvGHeoJ1zz~ngCKQ
[twitter]: https://twitter.com/OpenlineAI
