#!/usr/bin/env python3
"""
Enhanced Terminal UI Module for ShadowBox
Provides hacker-style, highly readable terminal output with animations and effects.
"""

import os
import sys
import time
import random
import threading
from datetime import datetime
from typing import Optional, List, Dict, Any

# Import settings functions
try:
    from meta_ops.settings import get_theme, set_theme
except ImportError:
    # Fallback if settings module is not available
    def get_theme():
        return 'hacker'
    def set_theme(theme_name):
        return True

class ColorTheme:
    """Color theme definitions"""
    
    THEMES = {
        'hacker': {
            'name': 'Classic Hacker',
            'header': '\033[38;5;196m',      # Bright red
            'accent': '\033[38;5;51m',       # Electric blue
            'border': '\033[38;5;129m',      # Electric purple
            'success': '\033[92m',           # Bright green
            'warning': '\033[93m',           # Bright yellow
            'error': '\033[91m',             # Bright red
            'info': '\033[96m',              # Bright cyan
            'system': '\033[95m',            # Bright magenta
            'matrix': '\033[38;5;46m',       # Matrix green
            'subtitle': '\033[38;5;208m',    # Cyber orange
        },
        'matrix': {
            'name': 'Matrix Green',
            'header': '\033[38;5;46m',       # Matrix green
            'accent': '\033[38;5;40m',       # Darker green
            'border': '\033[38;5;34m',       # Forest green
            'success': '\033[92m',           # Bright green
            'warning': '\033[93m',           # Bright yellow
            'error': '\033[91m',             # Bright red
            'info': '\033[96m',              # Bright cyan
            'system': '\033[38;5;46m',       # Matrix green
            'matrix': '\033[38;5;46m',       # Matrix green
            'subtitle': '\033[38;5;40m',     # Darker green
        },
        'cyberpunk': {
            'name': 'Cyberpunk Neon',
            'header': '\033[38;5;201m',      # Hot pink
            'accent': '\033[38;5;51m',       # Electric blue
            'border': '\033[38;5;165m',      # Purple
            'success': '\033[92m',           # Bright green
            'warning': '\033[93m',           # Bright yellow
            'error': '\033[91m',             # Bright red
            'info': '\033[96m',              # Bright cyan
            'system': '\033[38;5;201m',      # Hot pink
            'matrix': '\033[38;5;51m',       # Electric blue
            'subtitle': '\033[38;5;165m',    # Purple
        },
        'ocean': {
            'name': 'Ocean Blue',
            'header': '\033[38;5;33m',       # Deep blue
            'accent': '\033[38;5;39m',       # Light blue
            'border': '\033[38;5;27m',       # Royal blue
            'success': '\033[92m',           # Bright green
            'warning': '\033[93m',           # Bright yellow
            'error': '\033[91m',             # Bright red
            'info': '\033[96m',              # Bright cyan
            'system': '\033[38;5;33m',       # Deep blue
            'matrix': '\033[38;5;39m',       # Light blue
            'subtitle': '\033[38;5;27m',     # Royal blue
        },
        'fire': {
            'name': 'Fire Red',
            'header': '\033[38;5;196m',      # Bright red
            'accent': '\033[38;5;202m',      # Orange red
            'border': '\033[38;5;208m',      # Orange
            'success': '\033[92m',           # Bright green
            'warning': '\033[93m',           # Bright yellow
            'error': '\033[91m',             # Bright red
            'info': '\033[96m',              # Bright cyan
            'system': '\033[38;5;196m',      # Bright red
            'matrix': '\033[38;5;202m',      # Orange red
            'subtitle': '\033[38;5;208m',    # Orange
        }
    }

class Colors:
    """ANSI color codes for terminal styling"""
    # Basic colors
    BLACK = '\033[30m'
    RED = '\033[31m'
    GREEN = '\033[32m'
    YELLOW = '\033[33m'
    BLUE = '\033[34m'
    MAGENTA = '\033[35m'
    CYAN = '\033[36m'
    WHITE = '\033[37m'
    
    # Bright colors
    BRIGHT_BLACK = '\033[90m'
    BRIGHT_RED = '\033[91m'
    BRIGHT_GREEN = '\033[92m'
    BRIGHT_YELLOW = '\033[93m'
    BRIGHT_BLUE = '\033[94m'
    BRIGHT_MAGENTA = '\033[95m'
    BRIGHT_CYAN = '\033[96m'
    BRIGHT_WHITE = '\033[97m'
    
    # Background colors
    BG_BLACK = '\033[40m'
    BG_RED = '\033[41m'
    BG_GREEN = '\033[42m'
    BG_YELLOW = '\033[43m'
    BG_BLUE = '\033[44m'
    BG_MAGENTA = '\033[45m'
    BG_CYAN = '\033[46m'
    BG_WHITE = '\033[47m'
    
    # Text styles
    BOLD = '\033[1m'
    DIM = '\033[2m'
    ITALIC = '\033[3m'
    UNDERLINE = '\033[4m'
    BLINK = '\033[5m'
    REVERSE = '\033[7m'
    STRIKETHROUGH = '\033[9m'
    
    # Reset
    RESET = '\033[0m'
    
    # Custom combinations
    MATRIX_GREEN = '\033[38;5;46m'
    NEON_BLUE = '\033[38;5;51m'
    ELECTRIC_PURPLE = '\033[38;5;129m'
    CYBER_ORANGE = '\033[38;5;208m'
    HACKER_RED = '\033[38;5;196m'

class Symbols:
    """Unicode symbols for enhanced visual appeal"""
    # Status symbols
    SUCCESS = '✓'
    FAIL = '✗'
    WARNING = '⚠'
    INFO = 'ℹ'
    DOWNLOAD = '⬇'
    UPLOAD = '⬆'
    SEARCH = '🔍'
    MUSIC = '♪'
    FOLDER = '📁'
    FILE = '📄'
    
    # Hacker-style symbols
    SKULL = '☠'
    LIGHTNING = '⚡'
    GEAR = '⚙'
    LOCK = '🔒'
    UNLOCK = '🔓'
    TERMINAL = '💻'
    SATELLITE = '📡'
    PALETTE = '🎨'
    
    # Progress symbols
    ARROW_RIGHT = '→'
    ARROW_LEFT = '←'
    ARROW_UP = '↑'
    ARROW_DOWN = '↓'
    DOUBLE_ARROW = '⇒'
    
    # Decorative
    DIAMOND = '◆'
    SQUARE = '■'
    CIRCLE = '●'
    TRIANGLE = '▲'
    STAR = '★'

class TerminalUI:
    """Enhanced terminal UI with hacker-style aesthetics"""
    
    def __init__(self, enable_animations: bool = True, enable_sound: bool = False, theme: str = None, font_scale: float = 1.1):
        self.enable_animations = enable_animations
        self.enable_sound = enable_sound
        self.terminal_width = self._get_terminal_width()
        self.loading_chars = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏']
        self.glitch_chars = ['█', '▓', '▒', '░', '▄', '▀', '■', '□', '▪', '▫', '╬', '╫', '╪', '┼', '┬', '┴', '├', '┤', '│', '─']
        self.static_chars = ['*', '#', '@', '%', '&', '+', '=', '~', '^', ':', ';', '<', '>', '?', '!', '|', '\\', '/', '-', '_']
        self.font_scale = font_scale
        
        # Load theme from settings if not specified
        if theme is None:
            theme = get_theme()
        
        self.theme_name = theme
        self.theme = ColorTheme.THEMES.get(theme, ColorTheme.THEMES['hacker'])
        
        # Apply font scaling if supported
        self._apply_font_scaling()
    
    def _apply_font_scaling(self):
        """Apply font scaling using terminal escape sequences"""
        if self.font_scale != 1.0:
            # Some terminals support font size changes via escape sequences
            # This is experimental and may not work in all terminals
            try:
                scale_code = f"\033]50;size={int(12 * self.font_scale)}\007"
                print(scale_code, end='', flush=True)
            except:
                pass  # Silently fail if not supported
    
    def set_theme(self, theme_name: str):
        """Change the color theme and save to settings"""
        if theme_name in ColorTheme.THEMES:
            self.theme_name = theme_name
            self.theme = ColorTheme.THEMES[theme_name]
            # Save theme to settings
            set_theme(theme_name)
            return True
        return False
    
    def get_available_themes(self) -> List[str]:
        """Get list of available themes"""
        return list(ColorTheme.THEMES.keys())
        
    def _get_terminal_width(self) -> int:
        """Get terminal width, default to 80 if unable to determine"""
        try:
            return os.get_terminal_size().columns
        except:
            return 80
    
    def clear_screen(self, with_animation=False):
        """Clear the terminal screen with optional style"""
        os.system('cls' if os.name == 'nt' else 'clear')
        # Animation is now integrated into the menu/header display
    
    def _glitch_static_effect(self, duration: float = 1.0):
        """Glitch/static effect with random characters and colors"""
        if not self.enable_animations:
            return
            
        width = min(self.terminal_width, 80)
        height = 8
        
        # Create multiple phases of the glitch effect
        phases = int(duration * 15)  # 15 frames per second
        
        for phase in range(phases):
            # Clear previous lines
            if phase > 0:
                print(f"\033[{height}A", end='')  # Move cursor up
            
            # Generate glitch pattern
            for line in range(height):
                glitch_line = ''
                
                # Vary intensity based on phase for dynamic effect
                intensity = 0.15 + 0.1 * abs(phase % 10 - 5) / 5  # Oscillating intensity
                
                for col in range(width):
                    if random.random() < intensity:
                        # Choose character type based on randomness
                        if random.random() < 0.6:
                            char = random.choice(self.static_chars)
                            color = self.theme['accent']
                        else:
                            char = random.choice(self.glitch_chars)
                            color = self.theme['matrix']
                        
                        # Occasionally add error/warning colors for more chaos
                        if random.random() < 0.1:
                            color = self.theme['error']
                        elif random.random() < 0.05:
                            color = self.theme['warning']
                        
                        glitch_line += f"{color}{char}{Colors.RESET}"
                    else:
                        glitch_line += ' '
                
                print(f"\r{glitch_line}")
            
            # Variable sleep for more organic feel
            sleep_time = 0.04 + random.uniform(-0.01, 0.01)
            time.sleep(sleep_time)
        
        # Clear the glitch effect
        print(f"\033[{height}A", end='')  # Move cursor up
        for _ in range(height):
            print('\r' + ' ' * width)
        print(f"\033[{height}A", end='')  # Move cursor back up
    
    def print_header(self, with_startup_animation=False):
        """Print enhanced application header"""
        self.clear_screen()
        
        # ASCII art header with theme colors - centered
        header_lines = [
            "███████╗██╗  ██╗ █████╗ ██████╗  ██████╗ ██╗    ██╗██████╗  ██████╗ ██╗  ██╗",
            "██╔════╝██║  ██║██╔══██╗██╔══██╗██╔═══██╗██║    ██║██╔══██╗██╔═══██╗╚██╗██╔╝",
            "███████╗███████║███████║██║  ██║██║   ██║██║ █╗ ██║██████╔╝██║   ██║ ╚███╔╝ ",
            "╚════██║██╔══██║██╔══██║██║  ██║██║   ██║██║███╗██║██╔══██╗██║   ██║ ██╔██╗ ",
            "███████║██║  ██║██║  ██║██████╔╝╚██████╔╝╚███╔███╔╝██████╔╝╚██████╔╝██╔╝ ██╗",
            "╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝  ╚══╝╚══╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝"
        ]
        
        print()  # Add some space at the top
        
        if with_startup_animation and self.enable_animations:
            # Typewriter effect for header lines
            for line in header_lines:
                padding = (self.terminal_width - len(line)) // 2
                self._typewriter_line(f"{' ' * padding}{self.theme['header']}{Colors.BOLD}")
                self._typewriter_effect(f"{line}{Colors.RESET}")
                time.sleep(0.005)  # Brief pause between header lines
            print()  # Add space after header
            time.sleep(0.01)  # Pause before subtitle
        else:
            # Normal header display
            for line in header_lines:
                padding = (self.terminal_width - len(line)) // 2
                print(f"{' ' * padding}{self.theme['header']}{Colors.BOLD}{line}{Colors.RESET}")
            print()  # Add space after header
        
        # Animated subtitle only on startup
        subtitle = "ADVANCED AUDIO ACQUISITION SYSTEM"
        subtitle_with_symbols = f"[{Symbols.TERMINAL}] {subtitle} [{Symbols.TERMINAL}]"
        
        if with_startup_animation and self.enable_animations:
            # Center the subtitle before animation
            subtitle_padding = (self.terminal_width - len(subtitle_with_symbols)) // 2
            self._typewriter_line(' ' * subtitle_padding)
            # Slower typewriter effect for subtitle to make it more dramatic
            self._typewriter_effect(f"{self.theme['subtitle']}{Colors.BOLD}{subtitle_with_symbols}{Colors.RESET}", delay=0.015)
            time.sleep(0.1)  # Longer pause before system info for dramatic effect
        else:
            # Center the subtitle
            subtitle_padding = (self.terminal_width - len(subtitle_with_symbols)) // 2
            print(f"{' ' * subtitle_padding}{self.theme['subtitle']}{Colors.BOLD}{subtitle_with_symbols}{Colors.RESET}")
        
        # System info bar
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        system_info = f"{Colors.DIM}[{timestamp}] {self.theme['success']}SYSTEM ONLINE{Colors.DIM} | STATUS: {self.theme['success']}OPERATIONAL{Colors.RESET}"
        
        border = f"{self.theme['border']}{'═' * self.terminal_width}{Colors.RESET}"
        
        # System info always appears instantly - no typewriter effect
        print(border)
        print(f"{Colors.DIM}{system_info.center(self.terminal_width)}{Colors.RESET}")
        print(border)
        
        print()
    
    def _typewriter_effect(self, text: str, delay: float = 0.001):
        """Typewriter effect for text"""
        if not self.enable_animations:
            print(text)
            return
            
        for char in text:
            print(char, end='', flush=True)
            time.sleep(delay)
        print()
    
    def _typewriter_line(self, text: str, delay: float = 0.0005):
        """Typewriter effect for a single line without newline"""
        if not self.enable_animations:
            print(text, end='', flush=True)
            return
            
        for char in text:
            print(char, end='', flush=True)
            time.sleep(delay)
    
    def print_menu(self, with_typewriter=False):
        """Print enhanced main menu - typewriter only affects header, menu appears instantly"""
        menu_items = [
            (f"{Symbols.SEARCH} Search and download a song", "AUDIO_SCAN"),
            (f"{Symbols.DOWNLOAD} Download from URL (YouTube or Bandcamp)", "URL_EXTRACT"),
            (f"{Symbols.MUSIC} Download YouTube playlist", "PLAYLIST_SYNC"),
            (f"{Symbols.FILE} Batch download from a list", "BATCH_PROCESS"),
            (f"{Symbols.GEAR} Settings", "CONFIG_MODE"),
            (f"{Symbols.FOLDER} View downloaded songs", "ARCHIVE_VIEW"),
            (f"{Symbols.SKULL} Exit", "SYSTEM_HALT")
        ]
        
        # Calculate the maximum width needed for centering
        max_item_length = max(len(f"[{i}] {item} ({code})") for i, (item, code) in enumerate(menu_items, 1))
        menu_width = max(max_item_length + 4, 60)  # Add some padding, minimum 60 chars
        
        # Center the menu title
        title = f"[{Symbols.TERMINAL}] MAIN CONTROL INTERFACE [{Symbols.TERMINAL}]"
        title_padding = (self.terminal_width - len(title)) // 2
        
        # Menu always appears instantly - no typewriter effect for menu
        print(f"{' ' * title_padding}{self.theme['accent']}{Colors.BOLD}{title}{Colors.RESET}")
        
        # Create centered menu box
        menu_padding = (self.terminal_width - menu_width) // 2
        border_line = f"{' ' * menu_padding}{self.theme['border']}{'╔' + '═' * (menu_width - 2) + '╗'}{Colors.RESET}"
        print(border_line)
        
        for i, (item, code) in enumerate(menu_items, 1):
            # Use consistent styling for all options except exit (option 7)
            number_color = self.theme['success'] if i != 7 else self.theme['error']
            
            # Extract just the text part (remove the symbol)
            item_parts = item.split(' ', 1)
            symbol = item_parts[0]
            text = item_parts[1] if len(item_parts) > 1 else ""
            
            # Calculate spacing for alignment
            full_text = f"[{i}] {symbol} {text} ({code})"
            content_length = len(full_text)
            padding_needed = menu_width - content_length - 4  # 4 for borders and spaces
            
            # Consistent formatting: [number] symbol text (code)
            menu_line = f"{' ' * menu_padding}{self.theme['border']}║{Colors.RESET} {number_color}[{i}]{Colors.RESET} {self.theme['accent']}{symbol}{Colors.RESET} {Colors.WHITE}{text}{Colors.RESET}{' ' * padding_needed}{Colors.DIM}({code}){Colors.RESET} {self.theme['border']}║{Colors.RESET}"
            print(menu_line)
        
        bottom_border = f"{' ' * menu_padding}{self.theme['border']}{'╚' + '═' * (menu_width - 2) + '╝'}{Colors.RESET}"
        print(bottom_border)
        print()
    
    def success(self, message: str, tag: str = "SUCCESS"):
        """Print success message with enhanced styling"""
        timestamp = self._get_timestamp()
        formatted_msg = f"{self.theme['success']}[{Symbols.SUCCESS}]{Colors.RESET} {Colors.BOLD}[{tag}]{Colors.RESET} {Colors.WHITE}{message}{Colors.RESET} {Colors.DIM}{timestamp}{Colors.RESET}"
        print(formatted_msg)
    
    def error(self, message: str, tag: str = "ERROR"):
        """Print error message with enhanced styling"""
        timestamp = self._get_timestamp()
        formatted_msg = f"{self.theme['error']}[{Symbols.FAIL}]{Colors.RESET} {Colors.BOLD}[{tag}]{Colors.RESET} {Colors.WHITE}{message}{Colors.RESET} {Colors.DIM}{timestamp}{Colors.RESET}"
        print(formatted_msg)
    
    def warning(self, message: str, tag: str = "WARNING"):
        """Print warning message with enhanced styling"""
        timestamp = self._get_timestamp()
        formatted_msg = f"{self.theme['warning']}[{Symbols.WARNING}]{Colors.RESET} {Colors.BOLD}[{tag}]{Colors.RESET} {Colors.WHITE}{message}{Colors.RESET} {Colors.DIM}{timestamp}{Colors.RESET}"
        print(formatted_msg)
    
    def info(self, message: str, tag: str = "INFO"):
        """Print info message with enhanced styling"""
        timestamp = self._get_timestamp()
        formatted_msg = f"{self.theme['info']}[{Symbols.INFO}]{Colors.RESET} {Colors.BOLD}[{tag}]{Colors.RESET} {Colors.WHITE}{message}{Colors.RESET} {Colors.DIM}{timestamp}{Colors.RESET}"
        print(formatted_msg)
    
    def system(self, message: str, tag: str = "SYSTEM"):
        """Print system message with enhanced styling"""
        timestamp = self._get_timestamp()
        formatted_msg = f"{self.theme['system']}[{Symbols.TERMINAL}]{Colors.RESET} {Colors.BOLD}[{tag}]{Colors.RESET} {Colors.WHITE}{message}{Colors.RESET} {Colors.DIM}{timestamp}{Colors.RESET}"
        print(formatted_msg)
    
    def audio(self, message: str, tag: str = "AUDIO"):
        """Print audio-related message with enhanced styling"""
        timestamp = self._get_timestamp()
        formatted_msg = f"{self.theme['accent']}[{Symbols.MUSIC}]{Colors.RESET} {Colors.BOLD}[{tag}]{Colors.RESET} {Colors.WHITE}{message}{Colors.RESET} {Colors.DIM}{timestamp}{Colors.RESET}"
        print(formatted_msg)
    
    def download(self, message: str, tag: str = "DOWNLOAD"):
        """Print download message with enhanced styling"""
        timestamp = self._get_timestamp()
        formatted_msg = f"{self.theme['subtitle']}[{Symbols.DOWNLOAD}]{Colors.RESET} {Colors.BOLD}[{tag}]{Colors.RESET} {Colors.WHITE}{message}{Colors.RESET} {Colors.DIM}{timestamp}{Colors.RESET}"
        print(formatted_msg)
    
    def scan(self, message: str, tag: str = "SCAN"):
        """Print scan/search message with enhanced styling"""
        timestamp = self._get_timestamp()
        formatted_msg = f"{self.theme['accent']}[{Symbols.SEARCH}]{Colors.RESET} {Colors.BOLD}[{tag}]{Colors.RESET} {Colors.WHITE}{message}{Colors.RESET} {Colors.DIM}{timestamp}{Colors.RESET}"
        print(formatted_msg)
    
    def directory(self, message: str, tag: str = "DIR"):
        """Print directory operation message with enhanced styling"""
        timestamp = self._get_timestamp()
        formatted_msg = f"{self.theme['system']}[{Symbols.FOLDER}]{Colors.RESET} {Colors.BOLD}[{tag}]{Colors.RESET} {Colors.WHITE}{message}{Colors.RESET} {Colors.DIM}{timestamp}{Colors.RESET}"
        print(formatted_msg)
    
    def api(self, message: str, tag: str = "API"):
        """Print API-related message with enhanced styling"""
        timestamp = self._get_timestamp()
        formatted_msg = f"{self.theme['success']}[{Symbols.SATELLITE}]{Colors.RESET} {Colors.BOLD}[{tag}]{Colors.RESET} {Colors.WHITE}{message}{Colors.RESET} {Colors.DIM}{timestamp}{Colors.RESET}"
        print(formatted_msg)
    
    def _get_timestamp(self) -> str:
        """Get formatted timestamp"""
        return f"[{datetime.now().strftime('%H:%M:%S')}]"
    
    def progress_bar(self, current: int, total: int, prefix: str = "Progress", 
                    suffix: str = "Complete", length: int = 50):
        """Display an enhanced progress bar"""
        percent = (current / total) * 100
        filled_length = int(length * current // total)
        
        # Create gradient effect for progress bar
        bar_chars = []
        for i in range(length):
            if i < filled_length:
                if i < filled_length * 0.3:
                    bar_chars.append(f"{Colors.HACKER_RED}█{Colors.RESET}")
                elif i < filled_length * 0.7:
                    bar_chars.append(f"{Colors.CYBER_ORANGE}█{Colors.RESET}")
                else:
                    bar_chars.append(f"{Colors.BRIGHT_GREEN}█{Colors.RESET}")
            else:
                bar_chars.append(f"{Colors.DIM}░{Colors.RESET}")
        
        bar = ''.join(bar_chars)
        
        print(f'\r{Colors.BRIGHT_CYAN}[{Symbols.LIGHTNING}]{Colors.RESET} {prefix} |{bar}| {percent:.1f}% {suffix}', end='', flush=True)
        
        if current == total:
            print()  # New line when complete
    
    def loading_spinner(self, message: str = "Processing", duration: float = 2.0):
        """Display animated loading spinner"""
        if not self.enable_animations:
            print(f"{Colors.BRIGHT_YELLOW}[{Symbols.GEAR}]{Colors.RESET} {message}...")
            time.sleep(duration)
            return
        
        start_time = time.time()
        i = 0
        
        while time.time() - start_time < duration:
            spinner_char = self.loading_chars[i % len(self.loading_chars)]
            print(f'\r{Colors.BRIGHT_YELLOW}[{spinner_char}]{Colors.RESET} {Colors.WHITE}{message}...{Colors.RESET}', end='', flush=True)
            time.sleep(0.1)
            i += 1
        
        print(f'\r{Colors.BRIGHT_GREEN}[{Symbols.SUCCESS}]{Colors.RESET} {Colors.WHITE}{message} complete!{Colors.RESET}')
    
    def hacker_banner(self, text: str):
        """Display text in hacker-style banner"""
        border_char = '═'
        border_length = len(text) + 6
        
        print(f"{Colors.MATRIX_GREEN}{Colors.BOLD}")
        print(f"╔{border_char * border_length}╗")
        print(f"║   {text}   ║")
        print(f"╚{border_char * border_length}╝")
        print(f"{Colors.RESET}")
    
    def glitch_text(self, text: str, intensity: int = 3):
        """Apply glitch effect to text"""
        if not self.enable_animations:
            print(text)
            return
        
        glitch_chars = ['█', '▓', '▒', '░', '▄', '▀', '■', '□']
        
        for _ in range(intensity):
            glitched = list(text)
            for i in range(len(glitched)):
                if random.random() < 0.1:
                    glitched[i] = random.choice(glitch_chars)
            
            print(f'\r{Colors.HACKER_RED}{"".join(glitched)}{Colors.RESET}', end='', flush=True)
            time.sleep(0.1)
        
        print(f'\r{Colors.WHITE}{text}{Colors.RESET}')
    
    def section_divider(self, title: str = ""):
        """Print a stylized section divider"""
        if title:
            title_formatted = f" {title} "
            padding = (self.terminal_width - len(title_formatted)) // 2
            divider = f"{self.theme['border']}{'─' * padding}{Colors.BRIGHT_WHITE}{title_formatted}{self.theme['border']}{'─' * padding}{Colors.RESET}"
        else:
            divider = f"{self.theme['border']}{'─' * self.terminal_width}{Colors.RESET}"
        
        print(divider)
    
    def input_prompt(self, prompt: str, input_type: str = "INPUT") -> str:
        """Enhanced input prompt with styling"""
        formatted_prompt = f"{self.theme['accent']}[{Symbols.ARROW_RIGHT}]{Colors.RESET} {Colors.BOLD}[{input_type}]{Colors.RESET} {Colors.WHITE}{prompt}{Colors.RESET} {self.theme['accent']}>{Colors.RESET} "
        return input(formatted_prompt)
    
    def confirmation_prompt(self, message: str) -> bool:
        """Enhanced confirmation prompt"""
        response = self.input_prompt(f"{message} (y/N)", "CONFIRM")
        return response.lower() in ['y', 'yes']
    
    def display_metadata(self, metadata: Dict[str, Any]):
        """Display metadata in a formatted table"""
        self.section_divider("METADATA ANALYSIS")
        
        for key, value in metadata.items():
            if value:
                key_formatted = f"{Colors.BRIGHT_CYAN}{key.upper():<15}{Colors.RESET}"
                value_formatted = f"{Colors.WHITE}{value}{Colors.RESET}"
                print(f"{Colors.DIM}│{Colors.RESET} {key_formatted} {Colors.DIM}│{Colors.RESET} {value_formatted}")
        
        self.section_divider()
    
    def print_settings_menu(self, current_dir, use_spotify, audio_format, verbose_logging=False):
        """Print enhanced settings menu"""
        menu_items = [
            (f"{Symbols.FOLDER} Change music directory", "DIR_CONFIG"),
            (f"{Symbols.MUSIC} Toggle Spotify metadata", "SPOTIFY_TOGGLE"),
            (f"{Symbols.GEAR} Change audio format", "FORMAT_CONFIG"),
            (f"{Symbols.PALETTE} Change color theme", "THEME_CONFIG"),
            (f"{Symbols.DOWNLOAD} Update yt-dlp", "YTDLP_UPDATE"),
            (f"{Symbols.INFO} Toggle verbose logging", "VERBOSE_TOGGLE"),
            (f"{Symbols.ARROW_LEFT} Back to main menu", "MAIN_RETURN")
        ]
        
        # Calculate the maximum width needed for centering
        max_item_length = max(len(f"[{i}] {item} ({code})") for i, (item, code) in enumerate(menu_items, 1))
        menu_width = max(max_item_length + 4, 60)  # Add some padding, minimum 60 chars
        
        # Display current settings
        self.section_divider("CURRENT SETTINGS")
        print(f"{self.theme['accent']}{Symbols.FOLDER}{Colors.RESET} {Colors.WHITE}Music Directory:{Colors.RESET} {Colors.DIM}{current_dir}{Colors.RESET}")
        print(f"{self.theme['accent']}{Symbols.MUSIC}{Colors.RESET} {Colors.WHITE}Spotify Metadata:{Colors.RESET} {self.theme['success'] if use_spotify else self.theme['error']}{'Enabled' if use_spotify else 'Disabled'}{Colors.RESET}")
        print(f"{self.theme['accent']}{Symbols.GEAR}{Colors.RESET} {Colors.WHITE}Audio Format:{Colors.RESET} {Colors.DIM}{audio_format}{Colors.RESET}")
        print(f"{self.theme['accent']}{Symbols.INFO}{Colors.RESET} {Colors.WHITE}Verbose Logging:{Colors.RESET} {self.theme['success'] if verbose_logging else self.theme['error']}{'Enabled' if verbose_logging else 'Disabled'}{Colors.RESET}")
        print()
        
        # Center the menu title
        title = f"[{Symbols.GEAR}] SETTINGS CONFIGURATION [{Symbols.GEAR}]"
        title_padding = (self.terminal_width - len(title)) // 2
        
        print(f"{' ' * title_padding}{self.theme['accent']}{Colors.BOLD}{title}{Colors.RESET}")
        
        # Create centered menu box
        menu_padding = (self.terminal_width - menu_width) // 2
        border_line = f"{' ' * menu_padding}{self.theme['border']}{'╔' + '═' * (menu_width - 2) + '╗'}{Colors.RESET}"
        print(border_line)
        
        for i, (item, code) in enumerate(menu_items, 1):
            # Use consistent styling for all options except back (option 7)
            number_color = self.theme['success'] if i != 7 else self.theme['warning']
            
            # Extract just the text part (remove the symbol)
            item_parts = item.split(' ', 1)
            symbol = item_parts[0]
            text = item_parts[1] if len(item_parts) > 1 else ""
            
            # Calculate spacing for alignment
            full_text = f"[{i}] {symbol} {text} ({code})"
            content_length = len(full_text)
            padding_needed = menu_width - content_length - 4  # 4 for borders and spaces
            
            # Consistent formatting: [number] symbol text (code)
            menu_line = f"{' ' * menu_padding}{self.theme['border']}║{Colors.RESET} {number_color}[{i}]{Colors.RESET} {self.theme['accent']}{symbol}{Colors.RESET} {Colors.WHITE}{text}{Colors.RESET}{' ' * padding_needed}{Colors.DIM}({code}){Colors.RESET} {self.theme['border']}║{Colors.RESET}"
            print(menu_line)
        
        bottom_border = f"{' ' * menu_padding}{self.theme['border']}{'╚' + '═' * (menu_width - 2) + '╝'}{Colors.RESET}"
        print(bottom_border)
        print()

    def print_audio_format_menu(self, current_format):
        """Print enhanced audio format selection menu"""
        formats = [
            ('alac', 'Apple Lossless Audio Codec, highest quality (recommended)'),
            ('opus', 'High quality, small file size'),
            ('m4a', 'Good quality, compatible with Apple devices'),
            ('mp3', 'Universal compatibility, decent quality'),
            ('flac', 'Lossless audio, large file size'),
            ('wav', 'Uncompressed audio, very large file size'),
            ('aac', 'Good quality, small file size')
        ]
        
        menu_items = []
        for i, (fmt, desc) in enumerate(formats, 1):
            current_indicator = " (current)" if fmt == current_format else ""
            menu_items.append((f"{Symbols.MUSIC} {fmt.upper()}{current_indicator}", f"{desc}"))
        
        menu_items.append((f"{Symbols.GEAR} Custom format", "CUSTOM_FORMAT"))
        menu_items.append((f"{Symbols.ARROW_LEFT} Keep current format", "FORMAT_KEEP"))
        
        # Calculate the maximum width needed for centering
        max_item_length = max(len(f"[{i}] {item}") for i, (item, _) in enumerate(menu_items, 1))
        menu_width = max(max_item_length + 4, 80)  # Add some padding, minimum 80 chars for descriptions
        
        # Display current format
        self.section_divider("CURRENT AUDIO FORMAT")
        print(f"{self.theme['accent']}{Symbols.MUSIC}{Colors.RESET} {Colors.WHITE}Current Format:{Colors.RESET} {self.theme['success']}{current_format.upper()}{Colors.RESET}")
        print()
        
        # Center the menu title
        title = f"[{Symbols.MUSIC}] AUDIO FORMAT SELECTION [{Symbols.MUSIC}]"
        title_padding = (self.terminal_width - len(title)) // 2
        
        print(f"{' ' * title_padding}{self.theme['accent']}{Colors.BOLD}{title}{Colors.RESET}")
        
        # Create centered menu box
        menu_padding = (self.terminal_width - menu_width) // 2
        border_line = f"{' ' * menu_padding}{self.theme['border']}{'╔' + '═' * (menu_width - 2) + '╗'}{Colors.RESET}"
        print(border_line)
        
        for i, (item, desc) in enumerate(menu_items, 1):
            # Use different colors for different types
            if i <= 7:  # Format options
                number_color = self.theme['success']
            elif i == 8:  # Custom
                number_color = self.theme['accent']
            else:  # Keep current
                number_color = self.theme['warning']
            
            # Extract just the text part (remove the symbol)
            item_parts = item.split(' ', 1)
            symbol = item_parts[0]
            text = item_parts[1] if len(item_parts) > 1 else ""
            
            # Calculate spacing for alignment
            if i <= 7 or i == 8:  # Show description for formats and custom
                full_text = f"[{i}] {symbol} {text}"
                content_length = len(full_text)
                padding_needed = menu_width - content_length - len(desc) - 6  # 6 for borders, spaces, and separator
                
                menu_line = f"{' ' * menu_padding}{self.theme['border']}║{Colors.RESET} {number_color}[{i}]{Colors.RESET} {self.theme['accent']}{symbol}{Colors.RESET} {Colors.WHITE}{text}{Colors.RESET}{' ' * padding_needed}{Colors.DIM}{desc}{Colors.RESET} {self.theme['border']}║{Colors.RESET}"
            else:  # Keep current - no description
                full_text = f"[{i}] {symbol} {text}"
                content_length = len(full_text)
                padding_needed = menu_width - content_length - 4  # 4 for borders and spaces
                
                menu_line = f"{' ' * menu_padding}{self.theme['border']}║{Colors.RESET} {number_color}[{i}]{Colors.RESET} {self.theme['accent']}{symbol}{Colors.RESET} {Colors.WHITE}{text}{Colors.RESET}{' ' * padding_needed} {self.theme['border']}║{Colors.RESET}"
            
            print(menu_line)
        
        bottom_border = f"{' ' * menu_padding}{self.theme['border']}{'╚' + '═' * (menu_width - 2) + '╝'}{Colors.RESET}"
        print(bottom_border)
        print()

    def print_playlist_options_menu(self):
        """Print enhanced playlist options menu"""
        menu_items = [
            (f"{Symbols.MUSIC} Use Spotify metadata", "SPOTIFY_META"),
            (f"{Symbols.DOWNLOAD} Use basic metadata only", "BASIC_META")
        ]
        
        # Calculate the maximum width needed for centering
        max_item_length = max(len(f"[{i}] {item} ({code})") for i, (item, code) in enumerate(menu_items, 1))
        menu_width = max(max_item_length + 4, 60)  # Add some padding, minimum 60 chars
        
        # Center the menu title
        title = f"[{Symbols.MUSIC}] METADATA OPTIONS [{Symbols.MUSIC}]"
        title_padding = (self.terminal_width - len(title)) // 2
        
        print(f"{' ' * title_padding}{self.theme['accent']}{Colors.BOLD}{title}{Colors.RESET}")
        
        # Create centered menu box
        menu_padding = (self.terminal_width - menu_width) // 2
        border_line = f"{' ' * menu_padding}{self.theme['border']}{'╔' + '═' * (menu_width - 2) + '╗'}{Colors.RESET}"
        print(border_line)
        
        for i, (item, code) in enumerate(menu_items, 1):
            # Use consistent styling
            number_color = self.theme['success']
            
            # Extract just the text part (remove the symbol)
            item_parts = item.split(' ', 1)
            symbol = item_parts[0]
            text = item_parts[1] if len(item_parts) > 1 else ""
            
            # Calculate spacing for alignment
            full_text = f"[{i}] {symbol} {text} ({code})"
            content_length = len(full_text)
            padding_needed = menu_width - content_length - 4  # 4 for borders and spaces
            
            # Consistent formatting: [number] symbol text (code)
            menu_line = f"{' ' * menu_padding}{self.theme['border']}║{Colors.RESET} {number_color}[{i}]{Colors.RESET} {self.theme['accent']}{symbol}{Colors.RESET} {Colors.WHITE}{text}{Colors.RESET}{' ' * padding_needed}{Colors.DIM}({code}){Colors.RESET} {self.theme['border']}║{Colors.RESET}"
            print(menu_line)
        
        bottom_border = f"{' ' * menu_padding}{self.theme['border']}{'╚' + '═' * (menu_width - 2) + '╝'}{Colors.RESET}"
        print(bottom_border)
        print()

    def print_artist_list_menu(self, artists_data):
        """Print enhanced artist list menu with song counts"""
        if not artists_data:
            self.warning("No music found. Download some songs first!")
            return
        
        # Create menu items for artists
        menu_items = []
        for i, (artist, song_count) in enumerate(artists_data, 1):
            menu_items.append((f"{Symbols.MUSIC} {artist}", f"{song_count} SONGS"))
        
        menu_items.append((f"{Symbols.ARROW_LEFT} Back to main menu", "MAIN_RETURN"))
        
        # Calculate the maximum width needed for centering
        max_item_length = max(len(f"[{i}] {item} ({code})") for i, (item, code) in enumerate(menu_items, 1))
        menu_width = max(max_item_length + 4, 70)  # Add some padding, minimum 70 chars
        
        # Display summary
        self.section_divider("MUSIC LIBRARY")
        print(f"{self.theme['accent']}{Symbols.FOLDER}{Colors.RESET} {Colors.WHITE}Total Artists:{Colors.RESET} {self.theme['success']}{len(artists_data)}{Colors.RESET}")
        total_songs = sum(count for _, count in artists_data)
        print(f"{self.theme['accent']}{Symbols.MUSIC}{Colors.RESET} {Colors.WHITE}Total Songs:{Colors.RESET} {self.theme['success']}{total_songs}{Colors.RESET}")
        print()
        
        # Center the menu title
        title = f"[{Symbols.FOLDER}] ARTIST SELECTION [{Symbols.FOLDER}]"
        title_padding = (self.terminal_width - len(title)) // 2
        
        print(f"{' ' * title_padding}{self.theme['accent']}{Colors.BOLD}{title}{Colors.RESET}")
        
        # Create centered menu box
        menu_padding = (self.terminal_width - menu_width) // 2
        border_line = f"{' ' * menu_padding}{self.theme['border']}{'╔' + '═' * (menu_width - 2) + '╗'}{Colors.RESET}"
        print(border_line)
        
        for i, (item, code) in enumerate(menu_items, 1):
            # Use different colors for back option
            number_color = self.theme['success'] if i != len(menu_items) else self.theme['warning']
            
            # Extract just the text part (remove the symbol)
            item_parts = item.split(' ', 1)
            symbol = item_parts[0]
            text = item_parts[1] if len(item_parts) > 1 else ""
            
            # Calculate spacing for alignment
            full_text = f"[{i}] {symbol} {text} ({code})"
            content_length = len(full_text)
            padding_needed = menu_width - content_length - 4  # 4 for borders and spaces
            
            # Consistent formatting: [number] symbol text (code)
            menu_line = f"{' ' * menu_padding}{self.theme['border']}║{Colors.RESET} {number_color}[{i}]{Colors.RESET} {self.theme['accent']}{symbol}{Colors.RESET} {Colors.WHITE}{text}{Colors.RESET}{' ' * padding_needed}{Colors.DIM}({code}){Colors.RESET} {self.theme['border']}║{Colors.RESET}"
            print(menu_line)
        
        bottom_border = f"{' ' * menu_padding}{self.theme['border']}{'╚' + '═' * (menu_width - 2) + '╝'}{Colors.RESET}"
        print(bottom_border)
        print()

    def print_artist_songs_display(self, artist, songs_data):
        """Print enhanced display of songs for a specific artist"""
        self.section_divider(f"SONGS BY {artist.upper()}")
        
        if not songs_data:
            self.warning("No songs found for this artist.")
            return
        
        print(f"{self.theme['accent']}{Symbols.MUSIC}{Colors.RESET} {Colors.WHITE}Artist:{Colors.RESET} {self.theme['success']}{artist}{Colors.RESET}")
        print(f"{self.theme['accent']}{Symbols.FILE}{Colors.RESET} {Colors.WHITE}Total Songs:{Colors.RESET} {self.theme['success']}{len(songs_data)}{Colors.RESET}")
        print()
        
        # Create a table-like display
        header = f"{Colors.WHITE}{'#':<3} {'Song Title':<50} {'Format':<6} {'Size':<8} {'Date':<16}{Colors.RESET}"
        print(header)
        print(f"{self.theme['border']}{'─' * 85}{Colors.RESET}")
        
        for i, (filename, ext, size_mb, mod_time) in enumerate(songs_data, 1):
            # Truncate filename if too long
            display_name = filename[:47] + "..." if len(filename) > 50 else filename
            
            # Color code by format
            format_color = self.theme['success'] if ext.lower() in ['alac', 'flac'] else self.theme['accent']
            
            song_line = f"{self.theme['success']}{i:<3}{Colors.RESET} {Colors.WHITE}{display_name:<50}{Colors.RESET} {format_color}{ext.upper():<6}{Colors.RESET} {Colors.DIM}{size_mb:.1f}MB{Colors.RESET}  {Colors.DIM}{mod_time}{Colors.RESET}"
            print(song_line)
        
        print(f"{self.theme['border']}{'─' * 85}{Colors.RESET}")
        print()

    def theme_selection_menu(self):
        """Display theme selection menu and return selected theme"""
        self.clear_screen()
        
        themes = list(ColorTheme.THEMES.keys())
        
        # Create menu items for themes
        menu_items = []
        for theme_name in themes:
            theme_data = ColorTheme.THEMES[theme_name]
            current_indicator = " (current)" if theme_name == self.theme_name else ""
            menu_items.append((f"{Symbols.PALETTE} {theme_data['name']}{current_indicator}", theme_name.upper()))
        
        menu_items.append((f"{Symbols.ARROW_LEFT} Cancel", "THEME_CANCEL"))
        
        # Calculate the maximum width needed for centering
        max_item_length = max(len(f"[{i}] {item} ({code})") for i, (item, code) in enumerate(menu_items, 1))
        menu_width = max(max_item_length + 4, 60)  # Add some padding, minimum 60 chars
        
        # Center the menu title
        title = f"[{Symbols.PALETTE}] COLOR THEME SELECTION [{Symbols.PALETTE}]"
        title_padding = (self.terminal_width - len(title)) // 2
        
        print(f"{' ' * title_padding}{self.theme['accent']}{Colors.BOLD}{title}{Colors.RESET}")
        
        # Create centered menu box
        menu_padding = (self.terminal_width - menu_width) // 2
        border_line = f"{' ' * menu_padding}{self.theme['border']}{'╔' + '═' * (menu_width - 2) + '╗'}{Colors.RESET}"
        print(border_line)
        
        for i, (item, code) in enumerate(menu_items, 1):
            # Use different colors for cancel option
            number_color = self.theme['success'] if i != len(menu_items) else self.theme['warning']
            
            # Show theme preview for theme options
            if i <= len(themes):
                theme_data = ColorTheme.THEMES[themes[i-1]]
                preview = f"{theme_data['header']}█{theme_data['accent']}█{theme_data['border']}█{Colors.RESET} "
            else:
                preview = ""
            
            # Extract just the text part (remove the symbol)
            item_parts = item.split(' ', 1)
            symbol = item_parts[0]
            text = item_parts[1] if len(item_parts) > 1 else ""
            
            # Calculate spacing for alignment
            full_text = f"[{i}] {symbol} {text} ({code})"
            content_length = len(full_text) + (4 if preview else 0)  # Add preview length
            padding_needed = menu_width - content_length - 4  # 4 for borders and spaces
            
            # Consistent formatting: [number] symbol text (code)
            menu_line = f"{' ' * menu_padding}{self.theme['border']}║{Colors.RESET} {number_color}[{i}]{Colors.RESET} {preview}{self.theme['accent']}{symbol}{Colors.RESET} {Colors.WHITE}{text}{Colors.RESET}{' ' * padding_needed}{Colors.DIM}({code}){Colors.RESET} {self.theme['border']}║{Colors.RESET}"
            print(menu_line)
        
        bottom_border = f"{' ' * menu_padding}{self.theme['border']}{'╚' + '═' * (menu_width - 2) + '╝'}{Colors.RESET}"
        print(bottom_border)
        print()
        
        try:
            choice = self.input_prompt("Select a theme", "THEME")
            choice_num = int(choice)
            
            if choice_num == len(menu_items):  # Cancel option
                return None
            elif 1 <= choice_num <= len(themes):
                selected_theme = themes[choice_num - 1]
                if self.set_theme(selected_theme):
                    self.success(f"Theme changed to '{ColorTheme.THEMES[selected_theme]['name']}'")
                    return selected_theme
                else:
                    self.error("Failed to change theme")
                    return None
            else:
                self.error("Invalid selection")
                return None
        except ValueError:
            self.error("Invalid input. Please enter a number.")
            return None
    
    def exit_animation(self):
        """Display a hacker-style exit animation with terminal clearing"""
        if not self.enable_animations:
            self.system("Application shut down.")
            os.system('cls' if os.name == 'nt' else 'clear')
            return
        
        # Phase 1: Quick shutdown sequence
        self.system("Shutting down...")
        time.sleep(0.2)
        
        # Phase 2: Brief pause for dramatic effect
        time.sleep(0.4)
        
        # Phase 3: Final shutdown message
        print(f"\n{self.theme['success']}[{Symbols.SUCCESS}]{Colors.RESET} {Colors.BRIGHT_GREEN}SHADOWBOX TERMINATED{Colors.RESET}")
        time.sleep(0.3)
        
        # Final clear
        os.system('cls' if os.name == 'nt' else 'clear')

# Global instance for easy access - loads theme from settings
ui = TerminalUI()

# Convenience functions for backward compatibility
def clear_screen():
    ui.clear_screen()

def print_header(with_startup_animation=False):
    ui.print_header(with_startup_animation)

def print_menu(with_typewriter=False):
    ui.print_menu(with_typewriter)

def print_settings_menu(current_dir, use_spotify, audio_format, verbose_logging=False):
    ui.print_settings_menu(current_dir, use_spotify, audio_format, verbose_logging)

def print_audio_format_menu(current_format):
    ui.print_audio_format_menu(current_format)

def print_playlist_options_menu():
    ui.print_playlist_options_menu()

def print_artist_list_menu(artists_data):
    ui.print_artist_list_menu(artists_data)

def print_artist_songs_display(artist, songs_data):
    ui.print_artist_songs_display(artist, songs_data)

def success(message: str, tag: str = "SUCCESS"):
    ui.success(message, tag)

def error(message: str, tag: str = "ERROR"):
    ui.error(message, tag)

def warning(message: str, tag: str = "WARNING"):
    ui.warning(message, tag)

def info(message: str, tag: str = "INFO"):
    ui.info(message, tag)

def system(message: str, tag: str = "SYSTEM"):
    ui.system(message, tag)

def audio(message: str, tag: str = "AUDIO"):
    ui.audio(message, tag)

def download(message: str, tag: str = "DOWNLOAD"):
    ui.download(message, tag)

def scan(message: str, tag: str = "SCAN"):
    ui.scan(message, tag)

def directory(message: str, tag: str = "DIR"):
    ui.directory(message, tag)

def api(message: str, tag: str = "API"):
    ui.api(message, tag)

def exit_animation():
    ui.exit_animation()