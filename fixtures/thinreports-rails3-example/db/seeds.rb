# This file should contain all the record creation needed to seed the database with its default values.
# The data can then be loaded with the rake db:seed (or created alongside the db with db:setup).
#
# Examples:
#
#   cities = City.create([{ name: 'Chicago' }, { name: 'Copenhagen' }])
#   Mayor.create(name: 'Emanuel', city: cities.first)

Task.create([{name: 'Create specification', due_date: Date.today}, 
             {name: 'Create a beta version', due_date: Date.today + 1}, 
             {name: 'Deploy', due_date: Date.today + 2}])
