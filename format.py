import io
import os


def clean_csv(input_filename, output_filename):
    with open(input_filename, "r") as f:
        lines = f.readlines()

    if not lines:
        return

    header = lines[0].strip()
    # Filter out lines that match the header (except for the first occurrence)
    formatted_lines = [lines[0]] + [
        line for line in lines[1:] if line.strip() != header
    ]

    # Save the cleaned content
    with open(output_filename, "w") as f:
        f.writelines(formatted_lines)

    print(f"File formatted successfully and saved to {output_filename}")


# Example usage:
# clean_csv('your_input_file.csv', 'formatted_data.csv')

if __name__ == "__main__":
    dirs = [
        "compression",
        "graphproc",
        "thumbnail"
    ]

    for dir in dirs:
        for file in os.listdir(dir):
            file_path = os.path.join(dir, file)
            clean_csv(file_path, file_path)
