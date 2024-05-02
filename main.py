import itertools

# Read the contents of the tld-list.txt file into a list
with open('tld_list.txt', 'r') as file:
    tlds = [line.strip() for line in file]

# Generate combinations of TLDs using itertools.product
combinations = itertools.product(tlds, repeat=2)

# Define the output file name
output_file = 'tld_combinations.txt'

# Open the output file in write mode
with open(output_file, 'w') as outfile:
    # Iterate over each combination and output the concatenated TLDs
    for combination in combinations:
        if combination[0] != combination[1]:  # Skip if the same TLD is being concatenated
            outfile.write(''.join(combination) + '\n')  # Write the combination to the output file

print("Combinations saved to", output_file)