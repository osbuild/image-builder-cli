"""Utility functions for test suite."""
import subprocess
import time


def subprocess_retry_run(cmd, max_retries=3, timeout=300, **kwargs):
    """Run a subprocess command with retry logic and timeout.

    Args:
        cmd: Command to run (list of strings)
        max_retries: Maximum number of retry attempts (default: 3)
        timeout: Timeout in seconds for each attempt (default: 300s = 5min)
        **kwargs: Additional arguments to pass to subprocess.run()

    Returns:
        subprocess.CompletedProcess: The successful result

    Raises:
        subprocess.CalledProcessError: If all retry attempts fail
        subprocess.TimeoutExpired: If timeout occurs on all attempts
    """
    # Set defaults for subprocess.run if not provided
    check = kwargs.pop('check', True)
    capture_output = kwargs.pop('capture_output', True)
    text = kwargs.pop('text', True)

    cmd_str = ' '.join(cmd) if isinstance(cmd, list) else cmd
    last_exception = None

    for attempt in range(max_retries):
        try:
            result = subprocess.run(
                cmd, timeout=timeout, check=check,
                capture_output=capture_output, text=text, **kwargs)
            return result  # Success
        except subprocess.TimeoutExpired as e:
            last_exception = e
            if attempt < max_retries - 1:
                wait_time = 2 ** attempt  # Exponential backoff: 1s, 2s, 4s
                print(f"Timeout running '{cmd_str}' (attempt {attempt + 1}/{max_retries}). "
                      f"Retrying in {wait_time}s...")
                time.sleep(wait_time)
            else:
                raise subprocess.CalledProcessError(
                    1, e.cmd, f"Timeout after {max_retries} attempts"
                ) from last_exception
        except subprocess.CalledProcessError as e:
            last_exception = e
            if attempt < max_retries - 1:
                wait_time = 2 ** attempt  # Exponential backoff: 1s, 2s, 4s
                stderr = e.stderr if e.stderr else "(no error output)"
                print(f"Failed to run '{cmd_str}' (attempt {attempt + 1}/{max_retries}): {stderr}. "
                      f"Retrying in {wait_time}s...")
                time.sleep(wait_time)
            else:
                raise  # Re-raise the last exception

    # Should never reach here due to the raise in the else blocks above
    return None
