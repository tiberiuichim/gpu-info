# NVIDIA Driver Update — CUDA Runtime Compatibility

## Current State

| Component | Version |
|---|---|
| Driver | 580.159.03 |
| CUDA Runtime (nvidia-smi) | 13.0 |
| CUDA Toolkit (nvcc) | 13.2 |

The **CUDA Runtime** version reported by `nvidia-smi` is the *maximum CUDA version the installed driver supports*. The **CUDA Toolkit** version is the compiler you installed. They can differ because a newer toolkit can compile code that runs on an older driver (as long as the driver supports the toolkit's minimum CUDA version).

## Required Driver for CUDA 13.2

Per [CUDA Toolkit Release Notes](https://docs.nvidia.com/cuda/cuda-toolkit-release-notes/):

| CUDA Toolkit | Min Driver (Linux) |
|---|---|
| CUDA 13.0 | >= 580.65.06 |
| CUDA 13.2 GA | >= **595.45.04** |
| CUDA 13.2 Update 1 | >= **595.58.03** |

Current driver **580.159.03** is in the 580.x branch, so the highest CUDA runtime it supports is **13.0**.

To get CUDA 13.2 runtime support, upgrade to the **595.x driver branch** (>= 595.45.04).

## Upgrade Steps

```bash
# Option 1: Auto-install recommended driver
sudo ubuntu-drivers autoinstall

# Option 2: Install specific driver version
sudo apt install nvidia-driver-595

# Reboot after upgrade
sudo reboot
```

After rebooting, `nvidia-smi` should report `CUDA Version: 13.2` (or higher).

Sources: [CUDA Toolkit 13.2 Release Notes](https://docs.nvidia.com/cuda/cuda-toolkit-release-notes/)
