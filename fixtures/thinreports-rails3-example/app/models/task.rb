class Task < ActiveRecord::Base
  attr_accessible :done, :due_date, :name

  def state
    done? ? 'done' : 'yet'
  end
end
